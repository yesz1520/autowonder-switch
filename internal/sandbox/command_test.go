//go:build darwin || (linux && (amd64 || arm64))

package sandbox

import (
	"bytes"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Re-execute the test binary to probe the same network policy on both OSes.
func TestConnect(t *testing.T) {
	addr := os.Getenv("PROBE_ADDR")
	if addr == "" {
		t.Skip("sandbox child only")
	}
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err == nil {
		conn.Close()
		t.Fatal("sandbox connected to host listener")
	}
}

func TestWorkspacePolicy(t *testing.T) {
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/usr/bin/bwrap"); os.IsNotExist(err) {
			t.Skip("requires bubblewrap")
		}
	}
	work := t.TempDir()
	outside := filepath.Join(t.TempDir(), "private")
	if err := os.WriteFile(outside, []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(work, "link")); err != nil {
		t.Fatal(err)
	}
	extra := t.TempDir()
	if err := os.WriteFile(filepath.Join(extra, "skill"), []byte("skill"), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(write bool, env map[string]string, args ...string) (string, error) {
		var out bytes.Buffer
		w := Workspace{Directory: work, Writable: write, ReadOnlyPaths: []string{extra}}
		err := w.Run(t.Context(), env, args, &out)
		return out.String(), err
	}
	if msg, err := run(true, nil, "cat", filepath.Join(extra, "skill")); msg != "skill" || err != nil {
		t.Fatalf("read auxiliary directory: %q %v", msg, err)
	}
	if _, err := run(true, nil, "touch", filepath.Join(extra, "new")); err == nil {
		t.Fatal("auxiliary directory was writable")
	}
	t.Setenv("AW_PARENT_SECRET", "unavailable")
	got, err := run(true, map[string]string{"AW_VALUE": "a b"}, "sh", "-c", `printf '%s:%s' "$AW_VALUE" "$AW_PARENT_SECRET" > marker && cat marker`)
	if err != nil || got != "a b:" {
		t.Fatalf("workspace write: %q %v", got, err)
	}
	got, err = run(false, nil, "cat", "marker")
	if err != nil || got != "a b:" {
		t.Fatalf("workspace read: %q %v", got, err)
	}
	for _, tc := range []struct {
		write  bool
		script string
	}{{false, "echo changed > marker"}, {true, "echo changed > link"}, {true, "cat link"}} {
		if msg, err := run(tc.write, nil, "sh", "-c", tc.script); err == nil {
			t.Fatalf("%q unexpectedly succeeded: %s", tc.script, msg)
		}
	}
	if b, err := os.ReadFile(outside); err != nil || string(b) != "unchanged" {
		t.Fatal("outside file changed", err)
	}
	src, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	in, err := os.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	out, err := os.OpenFile(filepath.Join(work, "probe"), os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(out, in)
	if err = errors.Join(err, out.Close()); err != nil {
		t.Fatal(err)
	}
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	msg, err := run(false, map[string]string{"PROBE_ADDR": ln.Addr().String()}, "./probe", "-test.run=^TestConnect$")
	if err != nil {
		t.Fatalf("network probe: %s %v", msg, err)
	}
}
