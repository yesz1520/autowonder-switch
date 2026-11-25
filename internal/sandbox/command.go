// Package sandbox prepares isolated local commands for an executor workspace.
package sandbox

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Workspace struct {
	Directory     string
	Writable      bool
	ReadOnlyPaths []string
}

// Run preserves argv boundaries. Task variables are installed by env inside
// the sandbox; the launcher itself never receives task-controlled variables.
func (w Workspace) Run(ctx context.Context, env map[string]string, argv []string, output io.Writer) error {
	if len(argv) == 0 || argv[0] == "" {
		return errors.New("missing workspace command")
	}
	if !filepath.IsAbs(w.Directory) {
		return errors.New("workspace must be an absolute directory")
	}
	dir, err := filepath.EvalSymlinks(w.Directory)
	if err != nil {
		return err
	}
	if dir == "/" {
		return errors.New("workspace cannot be the filesystem root")
	}
	var readOnly []string
	for _, path := range w.ReadOnlyPaths {
		if !filepath.IsAbs(path) {
			return errors.New("read-only path must be absolute")
		}
		root, err := filepath.EvalSymlinks(path)
		if err != nil {
			return err
		}
		if root == "/" {
			return errors.New("cannot expose the filesystem root")
		}
		readOnly = append(readOnly, root)
	}
	scratch := "/tmp"
	if runtime.GOOS == "darwin" {
		temp, err := os.MkdirTemp("", "autowonder-workspace-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(temp)
		scratch, err = filepath.EvalSymlinks(temp)
		if err != nil {
			return err
		}
	}
	args := []string{"/usr/bin/env", "-i", "PATH=" + os.Getenv("PATH"), "HOME=" + scratch, "TMPDIR=" + scratch, "GIT_CONFIG_NOSYSTEM=1"}
	for key, value := range env {
		if key == "" || strings.ContainsAny(key, "=\x00") || strings.HasPrefix(key, "-") || strings.ContainsRune(value, 0) {
			return errors.New("invalid workspace environment variable")
		}
		args = append(args, key+"="+value)
	}
	// Resolve command names with the sandbox's PATH. A shell keeps a leading
	// dash or an equals sign in the executable name from becoming an env option.
	args = append(args, "/bin/sh", "-c", `exec "$@"`, "workspace")
	cmd, err := platformCommand(ctx, dir, scratch, w.Writable, readOnly, append(args, argv...))
	if err != nil {
		return err
	}
	cmd.Dir = dir
	cmd.Env = []string{"PATH=/usr/bin:/bin"}
	cmd.WaitDelay = 2 * time.Second
	cmd.Stdout, cmd.Stderr = output, output
	return cmd.Run()
}
