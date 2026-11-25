//go:build linux && (amd64 || arm64)

package sandbox

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

// Only the workspace and system tools enter the mount namespace. No host
// /run, HOME, or temporary directory is mounted. --disable-userns requires
// a non-setuid bubblewrap >= 0.8 and a host permitting user namespaces.
func platformCommand(ctx context.Context, dir, _ string, writable bool, readOnly, argv []string) (*exec.Cmd, error) {
	binaryPath := "/usr/bin/bwrap"
	if _, err := os.Stat(binaryPath); err != nil {
		return nil, errors.New("Pi workspace commands require /usr/bin/bwrap (bubblewrap >= 0.8)")
	}
	filter, err := socketFilter()
	if err != nil {
		return nil, err
	}
	args := []string{"--unshare-all", "--unshare-user", "--disable-userns", "--die-with-parent", "--new-session", "--cap-drop", "ALL"}
	for _, path := range []string{"/usr", "/bin", "/sbin", "/lib", "/lib64", "/etc/ld.so.cache"} {
		args = append(args, "--ro-bind-try", path, path)
	}
	mount := "--ro-bind"
	if writable {
		mount = "--bind"
	}
	args = append(args, "--proc", "/proc", "--dev", "/dev", "--tmpfs", "/tmp", mount, dir, dir)
	for _, path := range readOnly {
		args = append(args, "--ro-bind", path, path)
	}
	args = append(args, "--chdir", dir, "--seccomp", "0", "--")
	cmd := exec.CommandContext(ctx, binaryPath, append(args, argv...)...)
	cmd.Stdin = bytes.NewReader(filter)
	return cmd, nil
}

// A network namespace separates IP traffic. Also disallow socket creation
// (including Unix-domain sockets in the workspace) and io_uring. Reject
// foreign syscall ABIs before inspecting syscall numbers.
func socketFilter() ([]byte, error) {
	var arch uint32
	switch runtime.GOARCH {
	case "amd64":
		arch = 0xc000003e
	case "arm64":
		arch = 0xc00000b7
	default:
		return nil, errors.New("unsupported Linux workspace sandbox architecture")
	}
	var b []byte
	stmt := func(op uint16, yes, no byte, k uint32) {
		b = binary.LittleEndian.AppendUint16(b, op)
		b = append(b, yes, no)
		b = binary.LittleEndian.AppendUint32(b, k)
	}
	const (
		load   = 0x20
		equal  = 0x15
		ret    = 0x06
		kill   = 0x80000000
		denied = 0x50001
		allow  = 0x7fff0000
	)
	stmt(load, 0, 0, 4) // seccomp_data.arch
	stmt(equal, 1, 0, arch)
	stmt(ret, 0, 0, kill)
	stmt(load, 0, 0, 0)          // seccomp_data.nr
	stmt(0x35, 0, 1, 0x40000000) // x32 ABI
	stmt(ret, 0, 0, kill)
	for _, nr := range []uint32{syscall.SYS_SOCKET, syscall.SYS_SOCKETPAIR, 425 /* io_uring_setup */} {
		stmt(equal, 0, 1, nr)
		stmt(ret, 0, 0, denied)
	}
	stmt(ret, 0, 0, allow)
	return b, nil
}
