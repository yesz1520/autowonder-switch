// Package daemon supervises independently registered executor identities on one host.
package daemon

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"unicode"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/executor"
)

const CodexHarness = "codex"

type Profile struct {
	EventHook       aw.EventHook `json:"-"`
	platformBase    string
	MemoryMode      string `json:"memoryMode"`
	MemoryRevision  string `json:"memoryRevision,omitempty"`
	Hostname        string `json:"hostname,omitempty"`
	ReportedVersion string `json:"reportedVersion,omitempty"`
	VersionRevision string `json:"versionRevision,omitempty"`

	Harness  string `json:"harness"`
	Key      string `json:"key"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	ID       int64  `json:"executorId,string"`
	Capacity int    `json:"capacity"`
	Enabled  bool   `json:"enabled"`
	Token    string `json:"token,omitempty"`
}

func (p *Profile) validate() error {
	if p.MemoryMode == "" {
		p.MemoryMode = "platform"
	}
	if !executor.ValidMemoryMode(p.MemoryMode) {
		return errors.New("不支持的记忆模式")
	}
	if p.Harness == "" {
		p.Harness = CodexHarness
	}
	if p.Harness != CodexHarness {
		return errors.New("不支持的本机 harness")
	}
	p.Hostname = strings.TrimSpace(p.Hostname)
	p.ReportedVersion = strings.TrimSpace(p.ReportedVersion)
	for _, value := range []string{p.Hostname, p.ReportedVersion} {
		if len(value) > 255 || strings.IndexFunc(value, unicode.IsControl) >= 0 {
			return errors.New("主机名和上报版本最多 255 字节，且不能包含控制字符")
		}
	}
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" || len(p.Name) > 160 {
		return errors.New("请填写执行器名称（最多 160 字节）")
	}
	if p.Capacity < 1 || p.Capacity > 50 {
		return errors.New("并发数须在 1 到 50 之间")
	}
	u, err := url.Parse(strings.TrimSpace(p.URL))
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || strings.Contains(p.URL, "#") {
		return errors.New("请填写有效的平台地址")
	}
	switch u.Scheme {
	case "ws", "http":
		u.Scheme = "http"
	case "wss", "https":
		u.Scheme = "https"
	default:
		return errors.New("平台地址须使用 ws、wss、http 或 https")
	}
	u.Host = strings.ToLower(u.Host)
	if (u.Scheme == "http" && u.Port() == "80") || (u.Scheme == "https" && u.Port() == "443") {
		u.Host = strings.TrimSuffix(u.Host, ":"+u.Port())
	}
	basePath := strings.TrimSuffix(path.Clean("/"+u.Path), "/")
	if strings.HasSuffix(basePath, "/ws/executor") {
		basePath = strings.TrimSuffix(basePath, "/ws/executor")
	} else {
		basePath = strings.TrimSuffix(basePath, "/executors")
	}
	u.Path, u.RawPath = basePath, ""
	p.platformBase = u.String()
	if _, err := aw.New(aw.Config{BaseURL: p.platformBase, ExecutorID: p.ID, Token: p.Token, Capacity: p.Capacity}); err != nil {
		return err
	}
	u.Scheme = "ws" + strings.TrimPrefix(u.Scheme, "http")
	p.URL = u.JoinPath("ws/executor").String()
	return nil
}

// PlatformBaseURL is the HTTP base for API requests and existing executor state.
func (p Profile) PlatformBaseURL() string { return p.platformBase }

type View struct {
	RecordedVersion string `json:"recordedVersion,omitempty"`
	Profile
	Status string `json:"status"`
}
type Run func(context.Context, Profile, func(string)) error

type worker struct {
	cancel context.CancelFunc
	done   chan struct{}
	mu     sync.Mutex
	status string
}

func (w *worker) set(s string) { w.mu.Lock(); defer w.mu.Unlock(); w.status = s }
func (w *worker) get() string  { w.mu.Lock(); defer w.mu.Unlock(); return w.status }

type Manager struct {
	journalMu sync.Mutex
	mu        sync.Mutex
	ctx       context.Context
	file      string
	profiles  []Profile
	workers   map[string]*worker
	defaults  func(context.Context, string) (string, error)
	run       Run
	closed    bool
}

func Open(ctx context.Context, dir string, run Run) (*Manager, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	m := &Manager{ctx: ctx, file: filepath.Join(dir, "executors.json"), workers: map[string]*worker{}, run: run}
	b, err := os.ReadFile(m.file)
	if err == nil {
		err = json.Unmarshal(b, &m.profiles)
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	keys := map[string]bool{}
	for i := range m.profiles {
		p := &m.profiles[i]
		if err := p.validate(); err != nil {
			return nil, err
		}
		if p.Key == "" || keys[p.Key] || duplicate(m.profiles[:i], *p) {
			return nil, errors.New("duplicate or missing identity")
		}
		keys[p.Key] = true
	}
	for _, p := range m.profiles {
		if p.Enabled {
			m.start(p)
		}
	}
	return m, nil
}
func duplicate(profiles []Profile, p Profile) bool {
	return slices.ContainsFunc(profiles, func(v Profile) bool { return v.URL == p.URL && v.ID == p.ID })
}
func (m *Manager) start(p Profile) {
	ctx, cancel := context.WithCancel(m.ctx)
	w := &worker{cancel: cancel, done: make(chan struct{}), status: "connecting"}
	m.workers[p.Key] = w
	p.EventHook = func(e aw.Event) { m.recordEvent(p, e) }
	go func() {
		defer close(w.done)
		for ctx.Err() == nil {
			w.set("connecting")
			err := m.run(ctx, p, w.set)
			if errors.Is(err, executor.ErrRestartRequested) {
				continue
			}
			if ctx.Err() == nil {
				w.set("failed")
			}
			return
		}
	}()
}
func (m *Manager) stop(key string) {
	if w := m.workers[key]; w != nil {
		w.cancel()
		<-w.done
		delete(m.workers, key)
	}
}
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	for key := range m.workers {
		m.stop(key)
	}
}
func (m *Manager) list() []View {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]View, 0, len(m.profiles))
	for _, p := range m.profiles {
		p.MemoryMode = executor.SavedMemoryMode(filepath.Dir(m.file), p.platformBase, p.ID, p.MemoryMode, p.MemoryRevision)
		p.Token = ""
		s := "stopped"
		if w := m.workers[p.Key]; w != nil {
			s = w.get()
		}
		out = append(out, View{Profile: p, Status: s, RecordedVersion: executor.RecordedVersion(filepath.Dir(m.file), p.platformBase, p.ID, p.VersionRevision)})
	}
	return out
}

var errMissing = errors.New("未找到此执行器")
var errConflict = errors.New("请先停用执行器再修改连接或移除身份")
var errDuplicate = errors.New("这个平台的执行器 ID 已添加")

func (m *Manager) change(key string, p Profile, remove bool) (string, error) {
	return m.changeReporting(key, p, remove, false)
}
func (m *Manager) changeReporting(key string, p Profile, remove, reapply bool) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed || m.ctx.Err() != nil {
		return "", errors.New("daemon is stopping")
	}
	next := slices.Clone(m.profiles)
	i := slices.IndexFunc(next, func(v Profile) bool { return v.Key == key })
	if key != "" && i < 0 {
		return "", errMissing
	}
	if remove {
		if next[i].Enabled {
			return "", errConflict
		}
		next = slices.Delete(next, i, i+1)
	} else {
		var old Profile
		suppliedToken := p.Token != ""
		if i >= 0 {
			old = next[i]
			if p.Harness == "" {
				p.Harness = old.Harness
			}
			if p.Token == "" {
				p.Token = old.Token
			}
		}
		previousMode := old.MemoryMode
		if i >= 0 {
			previousMode = executor.SavedMemoryMode(filepath.Dir(m.file), old.platformBase, old.ID, old.MemoryMode, old.MemoryRevision)
			if p.MemoryMode == "" {
				p.MemoryMode = previousMode
			}
		}
		if err := p.validate(); err != nil {
			return "", err
		}
		memoryChanged := p.MemoryMode != previousMode
		p.MemoryRevision = old.MemoryRevision
		if memoryChanged {
			p.MemoryRevision = aw.NewResultID()
		}
		versionChanged := p.ReportedVersion != old.ReportedVersion || reapply
		p.VersionRevision = old.VersionRevision
		if versionChanged {
			p.VersionRevision = aw.NewResultID()
		}
		rebind := i >= 0 && (p.URL != old.URL || p.ID != old.ID)
		if rebind && !suppliedToken {
			return "", errors.New("更改平台或 ID 时请填写该身份的令牌")
		}
		if i >= 0 && old.Enabled && (memoryChanged || versionChanged || p.Hostname != old.Hostname || rebind || p.Harness != old.Harness || p.Token != old.Token || p.Capacity != old.Capacity) {
			return "", errConflict
		}
		if (i < 0 || rebind) && duplicate(next, p) {
			return "", errDuplicate
		}
		if i < 0 {
			var id [16]byte
			if _, err := rand.Read(id[:]); err != nil {
				return "", err
			}
			key = hex.EncodeToString(id[:])
		}
		p.Key = key
		if i < 0 {
			next = append(next, p)
		} else {
			next[i] = p
		}
	}
	if err := m.persist(next); err != nil {
		return "", fmt.Errorf("保存配置失败: %w", err)
	}
	m.profiles = next
	if remove || !p.Enabled {
		m.stop(key)
	} else if m.workers[key] == nil {
		m.start(p)
	}
	return key, nil
}
func (m *Manager) persist(next []Profile) error {
	f, err := os.CreateTemp(filepath.Dir(m.file), ".executors-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	err = json.NewEncoder(f).Encode(next)
	if err == nil {
		err = f.Sync()
	}
	if e := f.Close(); err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	return os.Rename(f.Name(), m.file)
}
