//go:build !darwin && !(linux && (amd64 || arm64))

package sandbox

import (
	"context"
	"errors"
	"os/exec"
)

func platformCommand(context.Context, string, string, bool, []string, []string) (*exec.Cmd, error) {
	return nil, errors.New("workspace sandbox requires macOS or Linux (amd64/arm64)")
}
