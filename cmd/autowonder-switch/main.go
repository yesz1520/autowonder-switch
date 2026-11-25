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
	"github.com/yesz1520/autowonder-switch/internal/daemon"
	"github.com/yesz1520/autowonder-switch/internal/executor"
	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/harness/codex"
)

func main() {
	admin := flag.String("listen", "127.0.0.1:8787", "local maintenance page address (multi-executor mode)")
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
	flag.Func("memory-mode", "single-executor memory mode: platform (default), provider-local, or none", func(mode string) error {
		if !executor.ValidMemoryMode(mode) {
			return fmt.Errorf("unknown memory mode %q", mode)
		}
		opts.MemoryMode, opts.MemoryRevision = mode, mode
		return nil
	})
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
	if *base == "" && *id == 0 {
		if opts.MemoryMode != "" {
			log.Fatal("set memory mode per executor in the maintenance page")
		}
		backends := map[string]harness.Backend{daemon.CodexHarness: backend}
		defaults := func(ctx context.Context, field string) (string, error) {
			switch field {
			case "hostname":
				return os.Hostname()
			case "reportedVersion":
				probe, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				v, err := backend.Version(probe)
				if err != nil {
					return "", err
				}
				return executor.DefaultCodexVersion(v), nil
			default:
				return "", errors.New("unknown field")
			}
		}
		err = daemon.Serve(ctx, *admin, *state, defaults, func(ctx context.Context, p daemon.Profile, status func(string)) error {
			options := opts
			options.ConnectionState = status
			options.MemoryMode, options.MemoryRevision = p.MemoryMode, p.MemoryRevision
			options.ReportedVersion, options.VersionRevision = p.ReportedVersion, p.VersionRevision
			reportedHost := p.Hostname
			if reportedHost == "" {
				var err error
				reportedHost, err = os.Hostname()
				if err != nil {
					return err
				}
			}

			cfg := aw.Config{BaseURL: p.PlatformBaseURL(), ExecutorID: p.ID, Token: p.Token, Capacity: p.Capacity, Hostname: reportedHost}
			selected, ok := backends[p.Harness]
			if !ok {
				return fmt.Errorf("unknown harness %q", p.Harness)
			}
			return executor.Run(ctx, cfg, *state, selected, repos, options)
		})
		if err != nil {
			log.Fatal(err)
		}
		return
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
