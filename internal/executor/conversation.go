package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
)

func (r *runtime) turn(ctx context.Context, s *aw.Session, q *aw.ConversationTurn) error {
	if q.ConversationID <= 0 || q.TurnID <= 0 {
		return errors.New("invalid conversation identity")
	}
	dir := filepath.Join(r.dir, fmt.Sprintf("conversation-%d", q.ConversationID))
	work := filepath.Join(dir, "workspace")
	if err := r.ensureWorkspace(dir); err != nil {
		return err
	}
	path := filepath.Join(dir, fmt.Sprintf("turn-%d-%d.json", q.TurnID, q.DispatchAttempt))
	ack := aw.ConversationTurnAck{ConversationID: q.ConversationID, TurnID: q.TurnID, DispatchAttempt: q.DispatchAttempt, ConversationDeliveryProtocolVersion: q.ConversationDeliveryProtocolVersion}
	if exists, err := r.loadFinal(path, &ack); err != nil {
		return err
	} else if exists {
		return r.send(&ack)
	}
	j, dup := r.reserve(r.turns, q.TurnID, q.ConversationID, q.DispatchAttempt)
	if dup {
		return nil
	}
	if j == nil {
		ack.Status = "FAILED"
		ack.Error = "conversation or executor busy"
		return r.send(&ack)
	}
	r.wg.Go(func() {
		defer r.finish(r.turns, q.TurnID)
		var reply harness.Reply
		err := markStart(path + ".started")
		if err == nil {
			reply, err = r.converse(j, q, dir, work)
		} else if errors.Is(err, errAlreadyStarted) {
			reply, err = r.recoverConversation(j.ctx, q, dir, work)
		}
		if r.ctx.Err() != nil {
			return
		}
		if reply.Usage != nil {
			ack.Usage, _ = json.Marshal(reply.Usage)
		}
		switch {
		case errors.Is(j.ctx.Err(), context.Canceled):
			ack.Status = "CANCELED"
		case err != nil:
			ack.Status = "FAILED"
			ack.Error = err.Error()
		default:
			ack.Status = "SUCCESS"
			ack.SessionID = reply.Session
			ack.ReplyMarkdown = reply.Text
		}
		if ack.ConversationDeliveryProtocolVersion == 1 {
			r.finishDelivery(path, &ack)
			return
		}
		if err = save(path, ack); err != nil {
			ack.Status = "FAILED"
			ack.Error = err.Error()
			ack.ReplyMarkdown = ""
		}
		r.report(&ack)
	})
	return nil
}
func (r *runtime) converse(j *job, q *aw.ConversationTurn, dir, workspace string) (harness.Reply, error) {
	spec := harness.Request{MemoryMode: r.launch.MemoryMode, Platform: r.platform(q.MCPToken), Environment: q.EnvironmentVariables, Directory: workspace, Prompt: q.Content, Instructions: q.SystemPrompt}
	var reply harness.Reply
	sessionFile := filepath.Join(dir, "session")
	b, err := os.ReadFile(sessionFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return reply, err
	}
	spec.Session = string(b)
	if q.CLISessionRef != "" && q.CLISessionRef != spec.Session {
		return reply, errors.New("Codex session does not belong to this conversation")
	}
	if q.CapabilityDownloadURL != "" {
		path := filepath.Join(dir, "capabilities.zip")
		if err = r.client.DownloadVerified(j.ctx, q.CapabilityDownloadURL, q.CapabilitySHA256, path); err != nil {
			return reply, err
		}
		m, err := aw.ReadTaskPackage(path)
		if err != nil {
			return reply, err
		}
		if m.PackageID != fmt.Sprintf("conversation-capabilities:%d:%d", q.ConversationID, q.TurnID) {
			return reply, errors.New("incorrect capability package identity")
		}
		text, err := packageText(path, false, &spec)
		if err != nil {
			return reply, err
		}
		spec.Instructions += "\n" + text
		closeTools, err := r.openPackageTools(j.ctx, path, q.MCPSecrets, &spec)
		if err != nil {
			return reply, err
		}
		defer closeTools()
		closeCapabilities, err := r.openPackageCapabilities(path, &spec)
		if err != nil {
			return reply, err
		}
		defer closeCapabilities()
		closeRepos, err := r.openConversationRepos(j.ctx, path, dir, &spec)
		if err != nil {
			return reply, err
		}
		defer closeRepos()
	}
	if err = r.attachments(j.ctx, q, &spec); err != nil {
		return reply, err
	}
	// Keep shell access for subsequent turns reading an earlier attachment.
	if _, err = os.Stat(filepath.Join(workspace, "attachments")); err == nil {
		spec.Writable = true
	}
	folder := fmt.Sprintf("outputs/%d-%d", q.TurnID, q.DispatchAttempt)
	var root *os.Root
	if r.files.Artifacts != nil {
		root, err = os.OpenRoot(workspace)
		if err != nil {
			return reply, err
		}
		defer root.Close()
		if err = root.MkdirAll(folder, 0700); err != nil {
			return reply, err
		}
		spec.Writable = true
		spec.Instructions += "\nPlace only files requested for delivery in " + folder + ". The executor attaches each file to your final reply."
	}
	events := &eventStream{r: r, q: q}
	spec.Ask = events.ask
	spec.Approvals = r.files.Approvals
	spec.Writable = spec.Writable || spec.Approvals
	events.emit("status", map[string]string{"status": "running"})
	reply, err = r.backend.Run(j.ctx, spec, harness.Hooks{TurnStarted: recordNativeTurn(filepath.Join(dir, fmt.Sprintf("turn-%d-%d.native.json", q.TurnID, q.DispatchAttempt))), Message: func(text string) { events.emit("text", map[string]string{"content": text}) }})

	if err != nil {
		return reply, err
	}
	return r.finishConversation(reply, root, folder, sessionFile)
}

func (r *runtime) recoverConversation(ctx context.Context, q *aw.ConversationTurn, dir, work string) (harness.Reply, error) {
	var native nativeTurn
	var reply harness.Reply
	ok, err := cached(filepath.Join(dir, fmt.Sprintf("turn-%d-%d.native.json", q.TurnID, q.DispatchAttempt)), &native)
	if err != nil {
		return reply, err
	}
	reader, supported := r.backend.(harness.TurnReader)
	if !ok || !supported {
		return reply, errAlreadyStarted
	}
	reply, status, err := reader.ReadTurn(ctx, work, native.Session, native.Turn)
	if err != nil {
		return reply, err
	}
	if status != "completed" || strings.TrimSpace(reply.Text) == "" {
		return reply, fmt.Errorf("previous turn status %s; no completed answer available", status)
	}
	var root *os.Root
	if r.files.Artifacts != nil {
		root, err = os.OpenRoot(work)
		if err != nil {
			return reply, err
		}
		defer root.Close()
	}
	return r.finishConversation(reply, root, fmt.Sprintf("outputs/%d-%d", q.TurnID, q.DispatchAttempt), filepath.Join(dir, "session"))
}

func (r *runtime) finishConversation(reply harness.Reply, root *os.Root, folder, sessionFile string) (harness.Reply, error) {
	if root != nil {
		links, e := r.files.Artifacts.publish(root, folder)
		if e != nil {
			return reply, e
		}
		if len(links) > 0 {
			reply.Text += "\n\n" + strings.Join(links, "\n\n")
		}
	}
	if err := writeFile(sessionFile, []byte(reply.Session)); err != nil {
		return reply, err
	}
	r.mu.Lock()
	if reply.Model != "" {
		r.model = reply.Model
	}
	r.mu.Unlock()
	return reply, nil
}
