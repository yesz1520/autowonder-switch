package codex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/yesz1520/autowonder-switch/internal/harness"
)

type frame struct {
	ID             json.RawMessage
	Method         string
	Params, Result json.RawMessage
	Error          *struct {
		Code    int
		Message string
	}
}
type rpc struct {
	cmd      *exec.Cmd
	cancel   context.CancelFunc
	enc      *json.Encoder
	write    sync.Mutex
	mu       sync.Mutex
	sequence int
	calls    map[string]chan frame
	events   chan frame
	done     chan struct{}
}

func start(ctx context.Context, binary, dir string, q harness.Request) (*rpc, error) {
	return startConfigured(ctx, binary, dir, q, nil)
}
func startConfigured(ctx context.Context, binary, dir string, q harness.Request, args []string) (*rpc, error) {
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, binary, append([]string{"app-server", "--stdio"}, args...)...)
	cmd.Dir = dir
	// Never pass the executor's platform credential to the child.
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "AUTOWONDER_TOKEN=") {
			cmd.Env = append(cmd.Env, v)
		}
	}
	cmd.WaitDelay = 2 * time.Second
	in, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		cancel()
		return nil, err
	}
	r := &rpc{cmd: cmd, cancel: cancel, enc: json.NewEncoder(in), calls: map[string]chan frame{}, events: make(chan frame, 128), done: make(chan struct{})}
	go r.read(ctx, out, q)
	if err = r.call(ctx, "initialize", map[string]any{"capabilities": map[string]bool{"experimentalApi": true}, "clientInfo": map[string]string{"name": "autowonder_switch", "version": "1"}}, nil); err == nil {
		err = r.send(map[string]string{"method": "initialized"})
	}
	if err != nil {
		r.close()
		return nil, err
	}
	return r, nil
}
func (r *rpc) close()           { r.cancel(); <-r.done }
func (r *rpc) send(v any) error { r.write.Lock(); defer r.write.Unlock(); return r.enc.Encode(v) }
func (r *rpc) read(ctx context.Context, out io.Reader, q harness.Request) {
	defer close(r.done)
	defer r.cmd.Wait()
	var workers sync.WaitGroup
	var pending sync.Map
	previousTool := make(chan struct{})
	close(previousTool)
	patches := map[string]json.RawMessage{}
	defer workers.Wait()
	defer r.cancel()
	dec := json.NewDecoder(out)
	for {
		var f frame
		if dec.Decode(&f) != nil {
			return
		}
		if len(f.ID) > 0 {
			if f.Method == "item/tool/call" {
				// Preserve tool order for stateful operations while allowing the
				// reader to process steering responses and cancellation notices.
				before := previousTool
				done := make(chan struct{})
				previousTool = done
				wait, stop := context.WithCancel(ctx)
				key := string(f.ID)
				pending.Store(key, stop)
				workers.Go(func() {
					defer close(done)
					defer pending.Delete(key)
					defer stop()
					select {
					case <-before:
					case <-ctx.Done():
						return
					}
					if wait.Err() != nil {
						return
					}
					text, err := platformCall(wait, q.Platform, f.Params)
					ok := err == nil
					if err != nil {
						text = err.Error()
					}
					if wait.Err() == nil {
						if err := r.send(map[string]any{"id": f.ID, "result": map[string]any{"success": ok, "contentItems": []map[string]string{{"type": "inputText", "text": text}}}}); err != nil {
							r.cancel()
						}
					}
				})
				continue
			}
			approval := f.Method == "item/commandExecution/requestApproval" || f.Method == "item/fileChange/requestApproval" || f.Method == "item/permissions/requestApproval"
			if q.Ask != nil && (f.Method == "item/tool/requestUserInput" || (q.Approvals && approval)) {
				if f.Method == "item/fileChange/requestApproval" {
					var details map[string]json.RawMessage
					if json.Unmarshal(f.Params, &details) == nil && details != nil {
						if changes := patches[string(details["itemId"])]; changes != nil {
							details["changes"] = changes
							f.Params, _ = json.Marshal(details)
						}
					}
				}
				wait, stop := context.WithCancel(ctx)
				key := string(f.ID)
				pending.Store(key, stop)
				workers.Go(func() {
					defer pending.Delete(key)
					defer stop()
					v, complete, err := interaction(wait, f.Method, f.Params, q.Ask)
					response := map[string]any{"id": f.ID}
					if err != nil {
						response["error"] = map[string]any{"code": -32602, "message": err.Error()}
					} else {
						response["result"] = v
					}
					written := wait.Err() == nil && r.send(response) == nil && err == nil
					if complete != nil {
						complete(written)
					}
				})
				continue
			}
			if f.Method != "" {
				if r.send(map[string]any{"id": f.ID, "error": map[string]any{"code": -32601, "message": "interactive request unavailable"}}) != nil {
					return
				}
				continue
			}
			r.mu.Lock()
			ch := r.calls[string(f.ID)]
			r.mu.Unlock()
			if ch != nil {
				ch <- f
			}
			continue
		}
		switch f.Method {
		case "item/started", "item/fileChange/patchUpdated":
			if q.Approvals {
				var v struct {
					ItemID, Changes json.RawMessage
					Item            struct {
						ID, Changes json.RawMessage
						Type        string
					}
				}
				if json.Unmarshal(f.Params, &v) == nil {
					if v.Item.Type == "fileChange" {
						v.ItemID = v.Item.ID
						v.Changes = v.Item.Changes
					}
					if len(v.ItemID) > 0 && len(v.Changes) > 0 {
						patches[string(v.ItemID)] = v.Changes
					}
				}
			}
		case "serverRequest/resolved":
			var resolved struct{ RequestID json.RawMessage }
			if json.Unmarshal(f.Params, &resolved) == nil {
				if fn, ok := pending.Load(string(resolved.RequestID)); ok {
					fn.(context.CancelFunc)()
				}
			}
		case "item/completed", "turn/completed", "thread/tokenUsage/updated":
			if f.Method == "item/completed" && len(patches) > 0 {
				var v struct{ Item struct{ ID json.RawMessage } }
				if json.Unmarshal(f.Params, &v) == nil {
					delete(patches, string(v.Item.ID))
				}
			}
			select {
			case r.events <- f:
			case <-ctx.Done():
				return
			}
		}
	}
}
func (r *rpc) call(ctx context.Context, method string, params, out any) error {
	ch := make(chan frame, 1)
	r.mu.Lock()
	r.sequence++
	id := r.sequence
	key := fmt.Sprint(id)
	r.calls[key] = ch
	r.mu.Unlock()
	defer func() { r.mu.Lock(); delete(r.calls, key); r.mu.Unlock() }()
	if err := r.send(map[string]any{"id": id, "method": method, "params": params}); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.done:
		return errors.New("Codex app-server disconnected")
	case f := <-ch:
		if f.Error != nil {
			return fmt.Errorf("Codex %s: %s", method, f.Error.Message)
		}
		if out != nil {
			return json.Unmarshal(f.Result, out)
		}
		return nil
	}
}
