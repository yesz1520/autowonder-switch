package executor

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

type dispatchLog struct {
	mu        sync.Mutex
	data      bytes.Buffer
	limit     int64
	truncated bool
}

func newDispatchLog(q *aw.TaskDispatch) *dispatchLog {
	var v struct {
		Enabled  bool
		MaxBytes *int64
	}
	if json.Unmarshal(q.DebugLog, &v) != nil || !v.Enabled {
		return nil
	}
	maxBytes := int64(200 << 20)
	if v.MaxBytes != nil {
		maxBytes = min(maxBytes, *v.MaxBytes)
	}
	d := &dispatchLog{limit: maxBytes}
	d.event("dispatch.accepted", q.DispatchID)
	return d
}

// Values are numeric measurements or IDs. Prompts, command output, signed
// URLs and environment variables do not enter the diagnostic archive.
func (d *dispatchLog) event(name string, value int64) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.truncated {
		return
	}
	line, _ := json.Marshal(struct {
		At    time.Time `json:"at"`
		Event string    `json:"event"`
		Value int64     `json:"value"`
	}{time.Now().UTC(), name, value})
	if int64(d.data.Len()+len(line)+1) > d.limit {
		d.truncated = true
		return
	}
	d.data.Write(line)
	d.data.WriteByte('\n')
}

func (d *dispatchLog) upload(ctx context.Context, c *aw.Client, q *aw.TaskDispatch, result *aw.TaskResult) json.RawMessage {
	if d == nil {
		return nil
	}
	status := "FAILED"
	if result.Success {
		status = "SUCCEEDED"
	}
	d.event("dispatch."+status, q.DispatchID)
	d.mu.Lock()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write(d.data.Bytes())
	if e := gz.Close(); err == nil {
		err = e
	}
	truncated := d.truncated
	d.mu.Unlock()
	if err != nil || int64(buf.Len()) > d.limit {
		return json.RawMessage(`{"status":"FAILED","error":"debug archive exceeds byte limit or could not be encoded"}`)
	}
	input := aw.DebugLogUploadRequest{SizeBytes: int64(buf.Len()), SHA256: fmt.Sprintf("%x", sha256.Sum256(buf.Bytes())), Truncated: truncated, DispatchStatus: status}
	link, err := c.DebugLogUpload(ctx, q.DispatchID, input)
	if err == nil {
		err = c.PutDebugLog(ctx, link, bytes.NewReader(buf.Bytes()), input.SizeBytes)
	}
	channel := "DIRECT"
	if err != nil {
		channel = "RELAY"
		path := fmt.Sprintf("debug/agent-%d-%d.log.gz", q.AgentID, q.DispatchID)
		meta, _ := json.Marshal(map[string]any{"path": path, "sha256": input.SHA256, "truncated": truncated})
		_, err = c.UploadArtifacts(ctx, q.DispatchID, []aw.UploadFile{{Name: path, Reader: bytes.NewReader(buf.Bytes())}}, result.ResultID+"-debug", []json.RawMessage{meta})
	}
	meta := map[string]any{"status": "UPLOADED", "channel": channel, "sha256": input.SHA256, "sizeBytes": input.SizeBytes, "truncated": truncated}
	if err != nil {
		meta["status"] = "FAILED"
		meta["error"] = "debug log upload failed"
	}
	data, _ := json.Marshal(meta)
	return data
}
