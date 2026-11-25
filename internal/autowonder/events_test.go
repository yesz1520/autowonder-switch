package autowonder

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProtocolEventMetadata(t *testing.T) {
	var events []Event
	hook := EventHook(func(e Event) { events = append(events, e) })
	hook.frame("received", []byte(`{"type":"TASK_DISPATCH","dispatchId":42,"dispatchMcpToken":"private-fixture","environmentVariables":{"secret":"private-fixture"},"message":"private-fixture"}`), nil)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "private-fixture", 403) }))
	defer srv.Close()
	c, err := New(Config{BaseURL: srv.URL, ExecutorID: 1, Token: "private-fixture", EventHook: hook})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Environment(t.Context())
	if err == nil || len(events) != 2 || events[0].Operation != "TASK_DISPATCH" || events[0].DispatchID != 42 || events[1].Code != 403 || events[1].Outcome != "failed" {
		t.Fatal(events)
	}
	b, _ := json.Marshal(events)
	if strings.Contains(string(b), "private-fixture") {
		t.Fatal("credential in journal")
	}
}
