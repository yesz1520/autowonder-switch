package executor

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type guidanceReceipt struct {
	Digest string
	Status string
	Error  string
}

// The server allocates IDs independently for terminal steering and task guidance.
// Reserve the receipt before touching the live turn. After a crash an unfinished
// receipt cannot establish whether Codex accepted the request.
func (r *runtime) applyOnce(ctx context.Context, kind string, dispatch, id int64, content string) (string, string) {
	if dispatch <= 0 || id <= 0 {
		return "FAILED", "missing dispatch or guidance ID"
	}
	dir := filepath.Join(r.dir, "guidance", fmt.Sprint(dispatch), kind)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "FAILED", err.Error()
	}
	path := filepath.Join(dir, fmt.Sprint(id))
	receipt := guidanceReceipt{Digest: fmt.Sprintf("%x", sha256.Sum256([]byte(content)))}
	initial, err := json.Marshal(receipt)
	if err != nil {
		return "FAILED", err.Error()
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if errors.Is(err, os.ErrExist) {
		data, err := os.ReadFile(path)
		var previous guidanceReceipt
		if err != nil || json.Unmarshal(data, &previous) != nil || previous.Status == "" {
			return "FAILED", "previous Codex steering outcome is unknown"
		}
		if previous.Digest != receipt.Digest {
			return "FAILED", "guidance ID reused with different content"
		}
		return previous.Status, previous.Error
	}
	if err != nil {
		return "FAILED", err.Error()
	}
	_, err = file.Write(initial)
	if err = errors.Join(err, file.Sync(), file.Close()); err != nil {
		return "FAILED", err.Error()
	}
	receipt.Status, receipt.Error = r.apply(ctx, dispatch, content)
	data, err := json.Marshal(receipt)
	if err == nil {
		err = writeFile(path, data)
	}
	if err != nil {
		return "FAILED", "could not persist steering outcome: " + err.Error()
	}
	return receipt.Status, receipt.Error
}
