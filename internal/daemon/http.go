package daemon

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

//go:embed ui/dist
var assets embed.FS

func Serve(ctx context.Context, addr, dir string, defaults func(context.Context, string, string) (string, error), run Run) error {
	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil || !a.IP.IsLoopback() {
		return errors.New("maintenance address must be loopback")
	}
	listener, err := net.ListenTCP("tcp", a)
	if err != nil {
		return err
	}
	defer listener.Close()
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	lock, err := os.OpenFile(filepath.Join(dir, "daemon.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return fmt.Errorf("cannot lock daemon state: %w", err)
	}
	owned, cancel := context.WithCancel(ctx)
	defer cancel()
	m, err := Open(owned, dir, run)
	if err != nil {
		return err
	}
	defer m.Close()
	m.defaults = defaults
	server := &http.Server{Handler: m.handler(listener.Addr().(*net.TCPAddr).Port), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second}
	stop := context.AfterFunc(owned, func() { server.Close() })
	defer stop()
	log.Printf("maintenance http://%s", listener.Addr())
	err = server.Serve(listener)
	cancel()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (m *Manager) handler(port int) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/executors/{key}/events", m.eventsHTTP)
	mux.HandleFunc("GET /api/executors", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m.list())
	})
	mux.HandleFunc("GET /api/defaults/{field}", func(w http.ResponseWriter, r *http.Request) {
		value, err := m.defaults(r.Context(), r.PathValue("field"), r.URL.Query().Get("harness"))
		if err != nil {
			http.Error(w, "无法读取该字段的本机默认值", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(value)
	})
	mutate := func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Profile
			ReapplyVersion bool `json:"reapplyVersion"`
		}
		if r.Method != http.MethodDelete {
			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "expected JSON", 415)
				return
			}
			d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10))
			d.DisallowUnknownFields()
			if err := d.Decode(&input); err != nil {
				http.Error(w, "无效的配置格式", 400)
				return
			}
		}
		key, err := m.changeReporting(r.PathValue("key"), input.Profile, r.Method == http.MethodDelete, input.ReapplyVersion)
		if err != nil {
			status := 400
			if errors.Is(err, errMissing) {
				status = 404
			}
			if errors.Is(err, errDuplicate) || errors.Is(err, errConflict) {
				status = 409
			}
			http.Error(w, err.Error(), status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"key": key})
	}
	mux.HandleFunc("POST /api/executors", mutate)
	mux.HandleFunc("PUT /api/executors/{key}", mutate)
	mux.HandleFunc("DELETE /api/executors/{key}", mutate)
	root, _ := fs.Sub(assets, "ui/dist")
	mux.Handle("GET /", http.FileServerFS(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, p, err := net.SplitHostPort(r.Host)
		if err != nil || p != fmt.Sprint(port) || (host != "localhost" && !net.ParseIP(host).IsLoopback()) {
			http.Error(w, "invalid host", 403)
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" && (r.Header.Get("Origin") != "http://"+r.Host || r.Header.Get("X-Switch-Request") != "local") {
			http.Error(w, "local origin required", 403)
			return
		}
		w.Header().Set("Content-Security-Policy", "default-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		mux.ServeHTTP(w, r)
	})
}
