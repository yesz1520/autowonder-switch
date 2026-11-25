package sandbox

import (
	"context"
	"os/exec"
	"strconv"
	"syscall"
)

const seatbelt = `(version 1)
(deny default)
(import "dyld-support.sb")
(allow signal (target self))
(allow process-fork process-exec)
(allow file-read-metadata sysctl-read)
(allow file-read* file-map-executable
  (subpath (param "WORKSPACE"))
  (subpath "/System/Library") (subpath "/Library/Developer")
  (subpath "/Applications/Xcode.app/Contents/Developer")
  (subpath "/usr") (subpath "/bin") (subpath "/sbin")
  (subpath "/opt/homebrew/bin") (subpath "/opt/homebrew/opt") (subpath "/opt/homebrew/Cellar") (subpath "/Library/Apple")
  (literal "/dev/null") (literal "/dev/zero") (literal "/dev/urandom")
  (literal "/dev/random") (literal "/private/etc/localtime"))
(allow file-read* file-write* file-map-executable (subpath (param "SCRATCH")))
(allow file-write-data (literal "/dev/null"))
`

func platformCommand(ctx context.Context, dir, scratch string, writable bool, readOnly, argv []string) (*exec.Cmd, error) {
	profile := seatbelt
	if writable {
		profile += `(allow file-write* (subpath (param "WORKSPACE")))`
	}
	args := []string{"-D", "WORKSPACE=" + dir, "-D", "SCRATCH=" + scratch}
	for i, path := range readOnly {
		key := "READ" + strconv.Itoa(i)
		args = append(args, "-D", key+"="+path)
		profile += `(allow file-read* file-map-executable (subpath (param "` + key + `")))`
	}
	args = append(args, "-p", profile)
	cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", append(args, argv...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	return cmd, nil
}
