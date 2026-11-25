package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/executor"
	"github.com/yesz1520/autowonder-switch/internal/harness/codex"
)

func main() {
	base := flag.String("url", "", "AutoWonder HTTP(S) base URL")
	id := flag.Int64("executor-id", 0, "registered executor ID")
	cap := flag.Int("capacity", 1, "maximum concurrent tasks and conversation turns")
	state := flag.String("state-dir", "", "persistent executor state directory")
	binary := flag.String("codex", "codex", "Codex executable")
	models := flag.Bool("list-models", false, "print selectable Codex models and exit")
	repos := map[string]string{}
	flag.Func("repo", "repo name=absolute local path (repeatable)", func(s string) error {
		name, path, ok := strings.Cut(s, "=")
		if !ok || name == "" || !filepath.IsAbs(path) {
			return fmt.Errorf("expected name=/absolute/path")
		}
		repos[name] = path
		return nil
	})
	var opts executor.FileOptions
	opts.RestartEnabled = runtime.GOOS == "darwin" || runtime.GOOS == "linux"
	flag.Func("attachment-origin", "trusted object storage origin for signed clarification attachments (repeatable)", func(v string) error {
		u, e := url.Parse(v)
		if e != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
			return fmt.Errorf("expected an HTTP(S) origin")
		}
		opts.AttachmentOrigins = append(opts.AttachmentOrigins, v)
		return nil
	})
	flag.BoolVar(&opts.NativePlugins, "native-package-plugins", false, "allow trusted AutoWonder packages to load native Codex plugins, including their executable components")
	flag.BoolVar(&opts.Approvals, "interactive-approvals", false, "allow the platform user to decide one-time native Codex approvals")
	listen := flag.String("artifact-listen", "127.0.0.1:0", "artifact download listener")
	public := flag.String("artifact-url", "", "external base URL forwarding /artifacts/ to the artifact listener")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	backend := codex.Backend{Binary: *binary}
	if *models {
		probe, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		cat, err := backend.Models(probe)
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range cat.Models {
			fmt.Printf("%s\t%s\n", v.ID, v.Name)
		}
		return
	}
	if *state == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			log.Fatal(err)
		}
		*state = filepath.Join(dir, "autowonder-switch")
	}
	if *public != "" {
		u, err := url.Parse(*public)
		if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") {
			log.Fatal("invalid artifact URL")
		}
		store := &executor.ArtifactStore{Dir: filepath.Join(*state, "published"), PublicURL: *public}
		ln, err := net.Listen("tcp", *listen)
		if err != nil {
			log.Fatal(err)
		}
		defer ln.Close()
		server := &http.Server{Handler: store, ReadHeaderTimeout: 5 * time.Second}
		defer server.Close()
		go server.Serve(ln)
		log.Printf("artifact listener %s", ln.Addr())
		opts.Artifacts = store
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	cfg := aw.Config{BaseURL: *base, ExecutorID: *id, Token: os.Getenv("AUTOWONDER_TOKEN"), Capacity: *cap, Hostname: hostname}
	if err = executor.Run(ctx, cfg, *state, backend, repos, opts); err != nil && ctx.Err() == nil {
		if errors.Is(err, executor.ErrRestartRequested) {
			path, e := os.Executable()
			if e != nil {
				log.Fatal(e)
			}
			err = syscall.Exec(path, os.Args, os.Environ())
		}
		log.Fatal(err)
	}
}
