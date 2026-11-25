package executor

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yesz1520/autowonder-switch/internal/taskhooks"
	"io/fs"
	"strings"
	"time"
	"unicode/utf16"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
)

type workflowStep struct {
	ID          json.Number `json:"id"`
	Name        string
	RetryBudget int
	GatePolicy  struct{ AskWhenUnclear, HumanApprovalRequired bool }
}

// A package can omit the workflow or include an empty workflow for legacy tasks.
func gatedSteps(pkg string, stepHooks bool) ([]workflowStep, error) {
	z, err := zip.OpenReader(pkg)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	f, err := z.Open("sdlc.json")
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var doc struct {
		CurrentStepID json.Number
		Steps         []workflowStep
	}
	if err = json.NewDecoder(f).Decode(&doc); err != nil {
		return nil, err
	}
	start := -1
	gated := false
	for i, s := range doc.Steps {
		if s.ID != "" && s.ID == doc.CurrentStepID {
			start = i
		}
		if start >= 0 && (s.GatePolicy.AskWhenUnclear || s.GatePolicy.HumanApprovalRequired) {
			gated = true
		}
	}
	if len(doc.Steps) > 0 && start < 0 {
		return nil, errors.New("workflow current step is missing")
	}
	if !gated && !stepHooks || start < 0 {
		return nil, nil
	}
	return doc.Steps[start:], nil
}

func retryHumanRequest(err error) bool {
	var status *aw.HTTPError
	if errors.As(err, &status) {
		return status.StatusCode == 408 || status.StatusCode == 429 || status.StatusCode >= 500
	}
	return errors.Is(err, aw.ErrPlatformTransport)
}

// The server owns expiration. An approval has no expiry and may wait longer
// than an INPUT form. Cancel the outstanding form if the owning turn stops.
func (r *runtime) humanRequest(ctx context.Context, dispatch int64, in aw.HumanRequestInput) (harness.Answer, error) {
	var answer harness.Answer
	var req aw.HumanRequest
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		req, err = r.client.CreateHumanRequest(ctx, dispatch, in)
		if err == nil || !retryHumanRequest(err) || attempt == 2 {
			break
		}
		select {
		case <-ctx.Done():
			return answer, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	if err != nil {
		return answer, err
	}
	id := req.ID
	if id <= 0 {
		return answer, errors.New("missing human request ID")
	}
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		if req.ID != id || req.DispatchID != dispatch || req.Key != in.Key || req.Kind != in.Kind {
			return answer, errors.New("human request identity mismatch")
		}
		switch req.Status {
		case "RESOLVED":
			var wire struct {
				Action  string
				Content map[string]any
			}
			if err = json.Unmarshal(req.Answer, &wire); err != nil {
				return answer, err
			}
			answer.Action, answer.Content = wire.Action, wire.Content
			valid := in.Kind == "INPUT" && (answer.Action == "accept" || answer.Action == "decline") || in.Kind == "APPROVAL" && (answer.Action == "approve" || answer.Action == "reject")
			if !valid {
				return answer, errors.New("invalid human request answer")
			}
			return answer, nil
		case "PENDING":
		case "EXPIRED", "CANCELED":
			return answer, fmt.Errorf("human request %s", req.Status)
		default:
			return answer, fmt.Errorf("unknown human request status %q", req.Status)
		}
		select {
		case <-ctx.Done():
			cleanup, stop := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			err = r.client.CancelHumanRequest(cleanup, dispatch, id)
			stop()
			return answer, errors.Join(ctx.Err(), err)
		case <-tick.C:
			next, e := r.client.GetHumanRequest(ctx, dispatch, id)
			if e != nil {
				if ctx.Err() != nil {
					continue
				}
				if !retryHumanRequest(e) {
					return answer, e
				}
				continue
			}
			req = next
		}
	}
}

func (r *runtime) runWorkflow(ctx context.Context, q *aw.TaskDispatch, pkg string, spec harness.Request, hooks harness.Hooks, repos []checkout, h *taskhooks.Set) (harness.Reply, error) {
	var reply harness.Reply
	steps, err := gatedSteps(pkg, h.Has("beforeStep") || h.Has("afterStep"))
	if err != nil {
		return reply, err
	}
	usage := taskUsage{}
	if len(steps) == 0 {
		reply, err = r.backend.Run(ctx, spec, hooks)
		step := ""
		if q.SDLCStepID > 0 {
			step = fmt.Sprint(q.SDLCStepID)
		}
		usage.report(ctx, r.client, q.DispatchID, step, reply.Usage)
		return reply, err
	}
	for i, step := range steps {
		if step.ID == "" {
			return reply, errors.New("missing workflow step ID")
		}
		current := spec
		current.Instructions += fmt.Sprintf("\nExecute only workflow step %s (%s) in this turn. The executor advances to the next step after its gate.", step.ID, step.Name)
		if i < len(steps)-1 {
			current.OutputSchema = nil
		}
		for attempt := 1; ; attempt++ {
			in := aw.HumanRequestInput{StepID: step.ID.String(), StepAttempt: attempt}
			if step.GatePolicy.AskWhenUnclear {
				current.Ask = func(ctx context.Context, msg string, schema map[string]any) (harness.Answer, error) {
					req := in
					req.Kind, req.Key, req.Message = "INPUT", aw.NewResultID(), msg
					data, err := json.Marshal(schema)
					if err != nil {
						return harness.Answer{}, err
					}
					req.Schema = data
					return r.humanRequest(ctx, q.DispatchID, req)
				}
			}
			fields := map[string]any{"stepId": step.ID.String(), "stepAttempt": attempt}
			if err = h.Run(ctx, "beforeStep", "", fields); err != nil {
				return reply, err
			}
			reply, err = r.backend.Run(ctx, current, hooks)
			usage.report(ctx, r.client, q.DispatchID, step.ID.String(), reply.Usage)
			if err == nil {
				err = h.Run(ctx, "afterStep", "", fields)
			}
			if err != nil {
				return reply, err
			}
			if !step.GatePolicy.HumanApprovalRequired {
				break
			}
			snapshot, err := repoCheckpoint(ctx, q.DispatchID, q.Attempt, "", reply.Session, reply.Text, repos)
			if err != nil {
				return reply, err
			}
			hash := fmt.Sprintf("%x", sha256.Sum256(snapshot))
			in.Kind, in.Key = "APPROVAL", fmt.Sprintf("approval:%s:%d:%s", step.ID, attempt, hash)
			in.Message = shortApproval(step.Name, reply.Text)
			in.Context, err = json.Marshal(map[string]string{"fingerprint": hash})
			if err != nil {
				return reply, err
			}
			answer, err := r.humanRequest(ctx, q.DispatchID, in)
			if err != nil {
				return reply, err
			}
			if answer.Action == "approve" {
				break
			}
			note, _ := answer.Content["note"].(string)
			if attempt > step.RetryBudget || strings.TrimSpace(note) == "" {
				return reply, fmt.Errorf("workflow approval rejected: %s", note)
			}
			current.Session, current.Fork = reply.Session, false
			current.Prompt = "Revise this step according to the review: " + note
		}
		spec.Session, spec.Fork = reply.Session, false
		spec.Prompt = "Continue with the next workflow step in the original work item."
	}
	return reply, nil
}

func shortApproval(name, text string) string {
	s := strings.TrimSpace(name + "\n" + text)
	units := utf16.Encode([]rune(s))
	if len(units) <= 1024 {
		return s
	}
	return string(utf16.Decode(units[:1023])) + "…"
}
