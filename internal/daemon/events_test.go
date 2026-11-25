package daemon

import (
	"encoding/json"
	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"net/http/httptest"
	"testing"
)

func TestEventHistoryBoundedAndIsolated(t *testing.T) {
	dir := t.TempDir()
	m, err := Open(t.Context(), dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	p := Profile{Name: "one", URL: "http://example.test", Capacity: 1, ID: 1, Token: "sample"}
	key, err := m.change("", p, false)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= recentEvents+4; i++ {
		m.recordEvent(m.profiles[0], aw.Event{DispatchID: int64(i)})
	}
	p.ID = 2
	other, err := m.change("", p, false)
	if err != nil {
		t.Fatal(err)
	}
	m.Close()
	m, err = Open(t.Context(), dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	for _, tc := range []struct {
		key         string
		code, count int
	}{{key, 200, recentEvents}, {other, 200, 0}, {"missing", 404, 0}} {
		req := httptest.NewRequest("GET", "http://localhost:8787/api/executors/"+tc.key+"/events", nil)
		w := httptest.NewRecorder()
		m.handler(8787).ServeHTTP(w, req)
		if w.Code != tc.code {
			t.Fatal(w.Code)
		}
		if w.Code != 200 {
			continue
		}
		var rows []aw.Event
		if err := json.Unmarshal(w.Body.Bytes(), &rows); err != nil || len(rows) != tc.count {
			t.Fatalf("rows: %d, error: %v", len(rows), err)
		}
		if len(rows) > 0 && (rows[0].DispatchID != recentEvents+4 || rows[len(rows)-1].DispatchID != 5) {
			t.Fatal("bad ordering")
		}
	}
}
