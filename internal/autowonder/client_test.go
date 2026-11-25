package autowonder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coder/websocket"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Exercise the published bilingual spec's English fixtures through the actual
// codecs. All payloads use concrete Go types, including negative and empty values.
func TestDocumentedFrames(t *testing.T) {
	models := []Outgoing{&Heartbeat{}, &TaskAck{}, &TaskBusy{}, &TaskProgress{}, &TaskPaused{}, &TaskPauseFailed{}, &TaskResult{}, &TaskGuidanceAck{}, &TaskSteerAck{}, &TaskHandoff{}, &ArtifactUploaded{}, &ConversationTurnEvent{}, &ConversationTurnAck{}, &ConversationCommandsResult{}, &ConversationElicitationReplyAck{}, &QoderModelCatalogResult{}, &MCPConnectionTestResult{}, &ExecutorRestartResult{}, &ExecutorUpgradeResult{}}
	out := map[string]reflect.Type{}
	for _, m := range models {
		out[m.Kind()] = reflect.TypeOf(m).Elem()
	}
	dir := "../../docs/protocol"
	files, err := filepath.Glob(filepath.Join(dir, "messages", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, filepath.Join(dir, "heartbeat.md"), filepath.Join(dir, "execution-inventory.md"))
	seen := map[string]bool{}
	examples := regexp.MustCompile("(?s)```json\n(.*?)\n```")
	for _, file := range files {
		if strings.HasSuffix(file, ".zh.md") {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range examples.FindAllSubmatch(data, -1) {
			var fields map[string]json.RawMessage
			if err = json.Unmarshal(m[1], &fields); err != nil {
				t.Fatal(err)
			}
			var kind string
			if err = json.Unmarshal(fields["type"], &kind); err != nil {
				t.Fatal(err)
			}
			seen[kind] = true
			var encoded []byte
			if typ, ok := out[kind]; ok {
				obj := reflect.New(typ).Interface().(Outgoing)
				if err = json.Unmarshal(m[1], obj); err == nil {
					encoded, err = Encode(obj)
				}
			} else {
				var obj Incoming
				obj, err = Decode(m[1])
				if _, unknown := obj.(*Unknown); unknown {
					t.Fatalf("missing decoder for %s", kind)
				}
				if err == nil {
					encoded, err = json.Marshal(obj)
					delete(fields, "type")
				}
			}
			if err != nil {
				t.Fatalf("%s: %v", file, err)
			}
			var got map[string]json.RawMessage
			if err = json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			for k, v := range fields {
				var wantValue, gotValue any
				if err = json.Unmarshal(v, &wantValue); err != nil {
					t.Fatal(err)
				}
				if wantValue == nil {
					continue
				}
				if err = json.Unmarshal(got[k], &gotValue); err != nil || !reflect.DeepEqual(gotValue, wantValue) {
					t.Errorf("%s %s.%s: got %s, want %s", file, kind, k, got[k], v)
				}
			}
		}
	}
	if len(seen) != 37 {
		t.Fatalf("covered %d frame types", len(seen))
	}
}

func TestSession(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	serverErr := make(chan error, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serve := func() error {
			q := r.URL.Query()
			if r.URL.Path != "/ws/executor" || q.Get("token") != "a&b" || q.Get("executorId") != "42" || q.Get("hostname") != "node" || q.Get("maxConcurrentDispatches") != "2" {
				return errors.New("bad handshake")
			}
			s, err := websocket.Accept(w, r, nil)
			if err != nil {
				return err
			}
			defer s.CloseNow()
			read := func(kind string) error {
				_, p, err := s.Read(ctx)
				if err != nil {
					return err
				}
				var h struct{ Type string }
				if err = json.Unmarshal(p, &h); err != nil {
					return err
				}
				if h.Type != kind {
					return fmt.Errorf("unexpected %s", h.Type)
				}
				return nil
			}
			if err := read("HEARTBEAT"); err != nil {
				return err
			}
			if err := read("HEARTBEAT"); err != nil {
				return err
			}
			if err := s.Write(ctx, websocket.MessageText, []byte(`{"type":"QODER_MODEL_CATALOG_REQUEST","executorId":42,"requestId":"abc","provider":"qoder"}`)); err != nil {
				return err
			}
			for {
				_, p, err := s.Read(ctx)
				if err != nil {
					return err
				}
				var result struct {
					Type, RequestID, Provider string
					Success                   bool
					Models                    []Model
				}
				if err = json.Unmarshal(p, &result); err != nil {
					return err
				}
				if result.Type == "HEARTBEAT" {
					continue
				}
				if result.Type != "QODER_MODEL_CATALOG_RESULT" || result.RequestID != "abc" || result.Provider != "qoder" || !result.Success || len(result.Models) != 1 {
					return errors.New("bad catalog response")
				}
				break
			}
			return s.Close(websocket.StatusPolicyViolation, "authentication failed")
		}
		serverErr <- serve()
	}))
	defer srv.Close()
	c, err := New(Config{BaseURL: srv.URL, ExecutorID: 42, Token: "a&b", Hostname: "node", Capacity: 2})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Serve(ctx, 20*time.Millisecond, func() Heartbeat { return Heartbeat{ProtocolFeatures: []Feature{FeatureQoderModelCatalogV1}} }, func(ctx context.Context, s *Session, in Incoming) error {
		q, ok := in.(*QoderModelCatalogRequest)
		if !ok {
			return errors.New("wrong message")
		}
		return s.Send(ctx, &QoderModelCatalogResult{RequestID: q.RequestID, Provider: q.Provider, Success: true, Models: []Model{{"model", "Model"}}})
	})
	if websocket.CloseStatus(err) != websocket.StatusPolicyViolation {
		t.Fatal(err)
	}
	if err := <-serverErr; err != nil {
		t.Fatal(err)
	}
}

func TestInventoryAndReceipts(t *testing.T) {
	ready := false
	h := Heartbeat{ProtocolFeatures: []Feature{FeatureDispatchInventoryV1}, MaxConcurrentDispatches: 2, DispatchInventoryReady: &ready, RunningDispatchIDs: []int64{}, OwnedDispatchIDs: []int64{}, RunningConversationTurnIDs: []int64{}}
	data, err := Encode(&h)
	if err != nil || !strings.Contains(string(data), `"runningDispatchIds":[]`) || !strings.Contains(string(data), `"dispatchInventoryReady":false`) {
		t.Fatalf("%s %v", data, err)
	}
	h.RunningDispatchIDs = []int64{1}
	if _, err = Encode(&h); err == nil {
		t.Fatal("unowned running ID allowed")
	}
	result := TaskResult{DispatchID: 10, ResultProtocolVersion: 2, ResultID: "a"}
	ack := TaskResultAck{DispatchID: 10, ResultProtocolVersion: 2, ResultID: "b"}
	if ack.Matches(&result) {
		t.Fatal("matched other result")
	}
	ack.ResultID = "a"
	if !ack.Matches(&result) {
		t.Fatal("missed matching result")
	}
}
