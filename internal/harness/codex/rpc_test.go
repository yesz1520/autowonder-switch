package codex

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yesz1520/autowonder-switch/internal/harness"
)

func TestSteerWhileToolIsRunning(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "app-server")
	script := `#!/bin/sh
read -r init
echo '{"id":1,"result":{}}'
read -r initialized
echo '{"method":"item/tool/call","id":11,"params":{"tool":"autowonder","arguments":{"operation":"call","name":"slow"}}}'
echo '{"method":"item/tool/call","id":12,"params":{"tool":"autowonder","arguments":{"operation":"call","name":"queued"}}}'
read -r steer
echo '{"method":"serverRequest/resolved","params":{"requestId":12}}'
echo '{"id":2,"result":{}}'
echo '{"method":"serverRequest/resolved","params":{"requestId":11}}'
read -r final
`
	if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	started, stopped := make(chan struct{}), make(chan struct{})
	request := harness.Request{Platform: func(ctx context.Context, name string, _ map[string]any) (string, error) {
		if name != "slow" {
			t.Errorf("canceled queued tool ran: %s", name)
			return "", nil
		}
		close(started)
		<-ctx.Done()
		close(stopped)
		return "", ctx.Err()
	}}
	r, err := start(ctx, binary, dir, request)
	if err != nil {
		t.Fatal(err)
	}
	defer r.close()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("tool did not start")
	}
	if err := r.call(ctx, "turn/steer", map[string]string{"threadId": "1"}, nil); err != nil {
		t.Fatal(err)
	}
	select {
	case <-stopped:
	case <-ctx.Done():
		t.Fatal("active tool did not receive cancellation")
	}
}
