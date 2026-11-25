package codex

import (
	"context"
	"errors"

	"github.com/yesz1520/autowonder-switch/internal/harness"
)

func (b Backend) ReadTurn(ctx context.Context, directory, session, turn string) (harness.Reply, string, error) {
	var reply harness.Reply
	r, err := start(ctx, b.Binary, directory, harness.Request{})
	if err != nil {
		return reply, "", err
	}
	defer r.close()
	var data struct {
		Thread struct {
			ID    string
			Turns []struct {
				ID, Status string
				Items      []struct{ Type, Text string }
			}
		}
	}
	if err = r.call(ctx, "thread/read", map[string]any{"threadId": session, "includeTurns": true}, &data); err != nil {
		return reply, "", err
	}
	if data.Thread.ID != session {
		return reply, "", errors.New("unexpected Codex thread in history")
	}
	for _, v := range data.Thread.Turns {
		if v.ID != turn {
			continue
		}
		reply.Session = session
		for _, item := range v.Items {
			if item.Type == "agentMessage" {
				reply.Text = item.Text
			}
		}
		return reply, v.Status, nil
	}
	return reply, "", errors.New("Codex turn missing from history")
}
