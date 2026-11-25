// Package executor translates platform messages into local harness operations.
package executor

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/yesz1520/autowonder-switch/internal/taskhooks"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
)

type job struct {
	packageHooks             *taskhooks.Set
	debug                    *dispatchLog
	repos                    []checkout
	stopModel                context.CancelFunc
	modelDone, pauseFinished chan struct{}
	control                  sync.Mutex
	native, latest           string
	portable, terminal       bool
	paused                   *aw.TaskPaused
	ctx                      context.Context
	cancel                   context.CancelFunc
	steer                    func(context.Context, string) error
	conversation, attempt    int64
}
type runtime struct {
	launch         launchConfig
	cleanups       chan *aw.WorkitemWorkspaceCleanup
	restartID      string
	restartHistory []string
	draining       string
	idle           chan struct{}
	shutdown       context.CancelCauseFunc
	nativeHooks    bool
	codexBinary    string
	executorID     int64
	upgrades       upgradeState
	questions      map[string]*pendingQuestion
	files          FileOptions
	*aw.Endpoint
	sources                            map[string]string
	ctx                                context.Context
	client                             *aw.Client
	backend                            harness.Backend
	pending                            *aw.Outbox
	dir, version, started, model, base string
	capacity                           int
	mu                                 sync.Mutex
	finals                             sync.Mutex
	delivery                           sync.Mutex
	session                            *aw.Session
	tasks, turns                       map[int64]*job
	wg                                 sync.WaitGroup
}

func identityDir(dir, base string, id int64) string {
	return filepath.Join(dir, fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s/%d", base, id)))))
}

func DefaultCodexVersion(version string) string { return "aws-codex-" + version }

// Run owns one registered identity. Background jobs survive WebSocket reconnects.
func Run(ctx context.Context, cfg aw.Config, dir string, b harness.Backend, sources map[string]string, opts ...FileOptions) error {
	if cfg.Capacity == 0 {
		cfg.Capacity = 1
	}
	c, err := aw.New(cfg)
	if err != nil {
		return err
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}
	owned, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	r := &runtime{shutdown: cancel, executorID: cfg.ExecutorID, base: cfg.BaseURL, sources: sources, client: c, ctx: owned, backend: b, capacity: cfg.Capacity, started: time.Now().UTC().Format(time.RFC3339Nano), tasks: map[int64]*job{}, turns: map[int64]*job{}}
	if b, ok := b.(interface{ Executable() string }); ok {
		r.codexBinary = b.Executable()
	}
	if len(opts) > 0 {
		r.files = opts[0]
	}
	if r.files.MemoryMode != "" && !ValidMemoryMode(r.files.MemoryMode) {
		return errors.New("invalid memory mode")
	}
	r.launch.MemoryMode = effectiveMemoryMode(restartState{}, r.files.MemoryMode, r.files.MemoryRevision)
	r.dir = identityDir(dir, cfg.BaseURL, cfg.ExecutorID)
	if r.pending, err = c.OpenOutbox(r.dir); err != nil {
		return err
	}
	if err = r.recoverFinalDeliveries(); err != nil {
		return err
	}
	if r.files.RestartEnabled {
		if err = r.loadRestart(); err != nil {
			return err
		}
		if r.launch.Capacity > 0 {
			cfg.Capacity = r.launch.Capacity
			r.capacity = cfg.Capacity
			c, err = aw.New(cfg)
			if err != nil {
				return err
			}
			r.client = c
		}
		if r.launch.DaemonName != "" {
			log.Printf("executor %d started as %q", r.executorID, r.launch.DaemonName)
		}
	}
	probe, stop := context.WithTimeout(ctx, 20*time.Second)
	defer stop()
	v, err := b.Version(probe)
	if err != nil {
		return err
	}
	r.version = DefaultCodexVersion(v)
	if err := r.loadUpgrade(); err != nil {
		return err
	}
	if cat, e := b.Models(probe); e == nil {
		r.model = cat.Default
	}
	if b, ok := b.(interface {
		SupportsToolHooks(context.Context, string) bool
	}); ok {
		r.nativeHooks = b.SupportsToolHooks(probe, r.dir)
	}
	r.Endpoint, err = aw.NewEndpoint(r.snapshot, r.pending)
	if err != nil {
		return err
	}
	if err = errors.Join(
		aw.On(r.Endpoint, r.task), aw.On(r.Endpoint, r.handoffResult), aw.On(r.Endpoint, r.queueCleanup),
		aw.On(r.Endpoint, r.turn), aw.On(r.Endpoint, r.answer),
		aw.On(r.Endpoint, r.cancelTurn),
		aw.On(r.Endpoint, r.catalog),
		aw.On(r.Endpoint, r.probeMCP),
		aw.On(r.Endpoint, r.steer),
		aw.On(r.Endpoint, r.pause),
		aw.On(r.Endpoint, r.guidance),
		aw.On(r.Endpoint, r.confirmPause),
		aw.On(r.Endpoint, r.commands),
		aw.On(r.Endpoint, r.upgrade), aw.On(r.Endpoint, r.restart),
	); err != nil {
		return err
	}
	r.cleanups = make(chan *aw.WorkitemWorkspaceCleanup, 32)
	r.wg.Go(r.runCleanups)
	err = c.Run(owned, 25*time.Second, r)
	cancel(nil)
	r.wg.Wait()
	if errors.Is(context.Cause(owned), ErrRestartRequested) {
		return ErrRestartRequested
	}
	return err
}
func (r *runtime) Connected(ctx context.Context, s *aw.Session) error {
	r.delivery.Lock()
	defer r.delivery.Unlock()
	r.session = s
	if err := r.flushUpgrade(ctx, s); err != nil {
		return err
	}
	if err := r.Endpoint.Connected(ctx, s); err != nil {
		return err
	}
	if err := r.replayPauses(ctx, s); err != nil {
		return err
	}
	// A RETRYABLE result or a lost receipt can occur on a healthy connection.
	r.wg.Go(func() {
		tick := time.NewTicker(time.Minute)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				if err := r.replayPauses(ctx, s); err != nil {
					return
				}
				if err := r.pending.Replay(ctx, s); err != nil {
					return
				}
			}
		}
	})
	return nil
}
func (r *runtime) send(m aw.Outgoing) error {
	r.delivery.Lock()
	defer r.delivery.Unlock()
	s := r.session
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return r.Endpoint.Send(ctx, s, m)
}
func (r *runtime) report(m aw.Outgoing) {
	if err := r.send(m); err != nil {
		log.Printf("send %s: %v", m.Kind(), err)
	}
}
func (r *runtime) snapshot() aw.Heartbeat {
	r.mu.Lock()
	defer r.mu.Unlock()
	tasks := make([]int64, 0, len(r.tasks))
	turns := make([]int64, 0, len(r.turns))
	for id := range r.tasks {
		tasks = append(tasks, id)
	}
	for id := range r.turns {
		turns = append(turns, id)
	}
	slices.Sort(tasks)
	slices.Sort(turns)
	waiting := map[int64]bool{}
	for _, question := range r.questions {
		waiting[question.q.TurnID] = true
	}
	computing := make([]int64, 0, len(turns))
	for _, id := range turns {
		if !waiting[id] {
			computing = append(computing, id)
		}
	}
	ready := true
	h := aw.Heartbeat{
		RestartRequestID: r.restartID, StartedAt: r.started, Version: r.version, Model: r.model,
		MaxConcurrentDispatches: r.capacity, DispatchInventoryReady: &ready,
		RunningDispatchIDs: tasks, OwnedDispatchIDs: tasks,
		RunningConversationTurnIDs: turns, ComputingConversationTurnIDs: computing,
		ProtocolFeatures: []aw.Feature{
			aw.FeatureDispatchInventoryV1, aw.FeatureConversationComputeInventoryV1, aw.FeatureWorkitemWorkspaceCleanup,
			aw.FeatureTaskModelSelectionV1, aw.FeatureAgentEnvironmentVariablesV1, aw.FeatureTaskPackageSignatureV1,
			aw.FeatureActionPlanV1, aw.FeatureTaskPackageHooksV1, aw.FeatureDebugLogV1, aw.FeatureDispatchRequestsV1, aw.FeatureExecutorUpdateV1, aw.FeatureQoderModelCatalogV1, aw.FeatureTerminalSteerV1, aw.FeatureCheckpointLightweightV1,
			aw.FeatureConversationAcpInteractionV1, aw.FeatureConversationAnswerDeliveryV1, aw.FeatureConversationTurnCancel, aw.FeatureConversationDeliveryAckV1, aw.FeatureConversationTurnEvent,
		},
	}
	if r.files.RestartEnabled {
		h.ProtocolFeatures = append(h.ProtocolFeatures, aw.FeatureExecutorRestartV1, aw.FeatureExecutorRestartPolicyV1, aw.FeatureExecutorLaunchCommandV1)
	}
	if r.nativeHooks {
		h.ProtocolFeatures = append(h.ProtocolFeatures, aw.FeatureTaskPackageToolHooksV1)
	}
	if len(r.files.AttachmentOrigins) > 0 {
		h.ProtocolFeatures = append(h.ProtocolFeatures, aw.FeatureAttachmentManifestV1)
	}
	if r.files.Artifacts != nil {
		h.ProtocolFeatures = append(h.ProtocolFeatures, aw.FeatureArtifactOutputV1)
	}
	return h
}
func (r *runtime) reserve(jobs map[int64]*job, id, conv, attempt int64) (*job, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if old := jobs[id]; old != nil && old.attempt == attempt && old.conversation == conv {
		return nil, true
	}
	if r.draining != "" || jobs[id] != nil || len(r.tasks)+len(r.turns) >= r.capacity {
		return nil, false
	}
	if conv != 0 {
		for _, other := range r.turns {
			if other.conversation == conv {
				return nil, false
			}
		}
	}
	ctx, cancel := context.WithCancel(r.ctx)
	j := &job{ctx: ctx, cancel: cancel, conversation: conv, attempt: attempt}
	jobs[id] = j
	return j, false
}
func (r *runtime) finish(jobs map[int64]*job, id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	jobs[id].cancel()
	delete(jobs, id)
	if r.idle != nil && len(r.tasks)+len(r.turns) == 0 {
		close(r.idle)
		r.idle = nil
	}
}
func (r *runtime) catalog(ctx context.Context, s *aw.Session, q *aw.QoderModelCatalogRequest) error {
	r.wg.Go(func() {
		probe, stop := context.WithTimeout(ctx, 15*time.Second)
		defer stop()
		cat, err := r.backend.Models(probe)
		m := &aw.QoderModelCatalogResult{RequestID: q.RequestID, Provider: q.Provider, Success: err == nil && len(cat.Models) > 0}
		for _, v := range cat.Models {
			m.Models = append(m.Models, aw.Model{ID: v.ID, Name: v.Name})
		}
		if ctx.Err() == nil {
			if e := s.Send(ctx, m); e != nil {
				log.Print(e)
			}
		}
	})
	return nil
}
func (r *runtime) steer(ctx context.Context, s *aw.Session, q *aw.TaskSteer) error {
	status, problem := r.applyOnce(ctx, q.Kind(), q.DispatchID, q.GuidanceID, q.Content)
	return s.Send(ctx, &aw.TaskSteerAck{DispatchID: q.DispatchID, GuidanceID: q.GuidanceID, Status: status, Error: problem})
}
func (r *runtime) guidance(ctx context.Context, s *aw.Session, q *aw.TaskGuidance) error {
	status, problem := "FAILED", "unsupported guidance mode"
	if q.Mode == "SAFE_STEER" {
		status, problem = r.applyOnce(ctx, q.Kind(), q.DispatchID, q.GuidanceID, q.Content)
	}
	return s.Send(ctx, &aw.TaskGuidanceAck{DispatchID: q.DispatchID, GuidanceID: q.GuidanceID, Status: status, Error: problem})
}
func (r *runtime) apply(ctx context.Context, id int64, text string) (string, string) {
	r.mu.Lock()
	var f func(context.Context, string) error
	if j := r.tasks[id]; j != nil {
		f = j.steer
	}
	r.mu.Unlock()
	if f == nil {
		return "FAILED", "no active Codex turn"
	}
	bounded, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	if err := f(bounded, text); err != nil {
		return "FAILED", err.Error()
	}
	return "APPLIED", ""
}

func (r *runtime) cancelTurn(_ context.Context, _ *aw.Session, q *aw.ConversationTurnCancel) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if j := r.turns[q.TurnID]; j != nil && j.conversation == q.ConversationID && j.attempt == q.DispatchAttempt {
		j.cancel()
	}
	return nil
}
func writeFile(path string, b []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".pending-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(f.Name(), path)
}

func (r *runtime) commands(ctx context.Context, s *aw.Session, q *aw.ConversationCommandsProbe) error {
	// Interactive TUI slash commands are not app-server operations.
	return s.Send(ctx, &aw.ConversationCommandsResult{ConversationID: q.ConversationID, Status: "OK", Commands: `{"availableCommands":[]}`})
}
