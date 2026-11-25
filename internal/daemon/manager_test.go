package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yesz1520/autowonder-switch/internal/executor"
)

func TestIndependentWorkersAndRestart(t *testing.T) {
	type invocation struct {
		key string
		end chan error
	}
	starts := make(chan invocation, 4)
	run := func(ctx context.Context, p Profile, state func(string)) error {
		end := make(chan error, 1)
		state("connected")
		starts <- invocation{p.Key, end}
		select {
		case e := <-end:
			return e
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	dir := t.TempDir()
	m, err := Open(t.Context(), dir, run)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	p := Profile{Name: "one", URL: "HTTP://LOCALHOST:80/", ID: 7, Token: "fixture-secret", Capacity: 1, Enabled: true}
	k, err := m.change("", p, false)
	if err != nil {
		t.Fatal(err)
	}
	next := func() invocation {
		t.Helper()
		select {
		case v := <-starts:
			return v
		case <-time.After(time.Second):
			t.Fatal("worker did not start")
			return invocation{}
		}
	}
	first := next()
	if first.key != k {
		t.Fatal("wrong worker")
	}
	if _, err = m.change("", p, false); !errors.Is(err, errDuplicate) {
		t.Fatal(err)
	}
	q := p
	q.ID = 8
	q.Name = "two"
	k2, err := m.change("", q, false)
	if err != nil {
		t.Fatal(err)
	}
	if next().key != k2 {
		t.Fatal("wrong second worker")
	}
	first.end <- executor.ErrRestartRequested
	if next().key != k {
		t.Fatal("restarted sibling")
	}
	p.Enabled = false
	p.Token = ""
	if _, err = m.change(k, p, false); err != nil {
		t.Fatal(err)
	}
	p.URL = "https://other.example"
	if _, err = m.change(k, p, false); err == nil {
		t.Fatal("reused a credential for another platform")
	}
	p.URL = q.URL
	p.ID = q.ID
	p.Token = "other-test"
	if _, err = m.change(k, p, false); !errors.Is(err, errDuplicate) {
		t.Fatal("allowed duplicate on edit", err)
	}
	rows := m.list()
	raw, _ := json.Marshal(rows)
	if strings.Contains(string(raw), "fixture-secret") || strings.Contains(string(raw), `"token"`) || rows[0].Status != "stopped" || rows[1].Status != "connected" {
		t.Fatal("incorrect public snapshot")
	}
	m.Close()
	restored, err := Open(t.Context(), dir, run)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	if next().key != k2 || restored.profiles[0].Token != "fixture-secret" {
		t.Fatal("incorrect restored configuration")
	}
	if _, err := restored.change(k, Profile{}, true); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(filepath.Join(dir, "executors.json"))
	if err != nil || st.Mode().Perm() != 0600 {
		t.Fatal("config permissions")
	}
}

func TestLocalManagementOrigin(t *testing.T) {
	m, err := Open(t.Context(), t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	h := m.handler(8787)
	for _, tc := range []struct {
		host, origin, header string
		want                 int
	}{
		{"localhost:8787", "http://unrelated.test", "local", 403},
		{"rebind.test:8787", "http://rebind.test:8787", "local", 403},
		{"localhost:8787", "http://localhost:8787", "", 403},
		{"localhost:8787", "http://localhost:8787", "local", 200},
	} {
		r := httptest.NewRequest("POST", "http://"+tc.host+"/api/executors", strings.NewReader(`{"name":"one","url":"http://example.test","executorId":"9007199254740993","capacity":1,"token":"test","enabled":false}`))
		r.Header.Set("Origin", tc.origin)
		r.Header.Set("X-Switch-Request", tc.header)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Errorf("got %d, want %d", w.Code, tc.want)
		}
	}
}

func TestExecutorURLAliases(t *testing.T) {
	for _, tc := range []struct{ input, endpoint, base string }{
		{"ws://example.test:7001/ws/executor", "ws://example.test:7001/ws/executor", "http://example.test:7001"},
		{"ws://example.test:7001", "ws://example.test:7001/ws/executor", "http://example.test:7001"},
		{"http://example.test:7001/executors", "ws://example.test:7001/ws/executor", "http://example.test:7001"},
		{"http://example.test:7001", "ws://example.test:7001/ws/executor", "http://example.test:7001"},
		{"https://EXAMPLE.test:443/proxy/executors/", "wss://example.test/proxy/ws/executor", "https://example.test/proxy"},
	} {
		p := Profile{Name: "alias", URL: tc.input, ID: 1, Token: "test", Capacity: 1}
		for range 2 {
			if err := p.validate(); err != nil || p.URL != tc.endpoint || p.PlatformBaseURL() != tc.base {
				t.Fatalf("%s: got %q, %q, %v", tc.input, p.URL, p.PlatformBaseURL(), err)
			}
		}
	}
}
