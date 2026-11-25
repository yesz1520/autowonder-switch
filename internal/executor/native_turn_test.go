package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
)

type completedTurn struct{ stub }

func (*completedTurn) ReadTurn(_ context.Context, _ string, session, turn string) (harness.Reply, string, error) {
	if session != "thread-a" || turn != "turn-b" {
		return harness.Reply{}, "", os.ErrNotExist
	}
	return harness.Reply{Session: session, Text: "recovered answer"}, "completed", nil
}
func TestRecoverCompletedConversation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "turn-3-2.native.json")
	if err := recordNativeTurn(path)("thread-a", "turn-b"); err != nil {
		t.Fatal(err)
	}
	backend := &completedTurn{}
	r := runtime{backend: backend}
	reply, err := r.recoverConversation(t.Context(), &aw.ConversationTurn{TurnID: 3, DispatchAttempt: 2}, dir, dir)
	if err != nil || reply.Text != "recovered answer" || backend.calls.Load() != 0 {
		t.Fatalf("reply=%v err=%v", reply, err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "session"))
	if err != nil || string(b) != "thread-a" {
		t.Fatalf("session=%q err=%v", b, err)
	}
}
