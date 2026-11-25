package executor

import (
	"context"
	"encoding/json"
	"fmt"
	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/harness/codex"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDispatchMemory(t *testing.T) {
	const storeID = 9007199254740993
	var recalls, writes atomic.Int32
	base := "/api/daemon/dispatches/71/memory/"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer dispatch-key" {
			t.Error("incorrect credential")
		}
		switch r.URL.Path {
		case base + "snapshot":
			fmt.Fprintf(w, `{"stores":[{"id":%d,"scope":"AGENT","permission":"WRITE","revision":3,"documents":[{"path":"topic.md","contentMd":"Passphrase: quartz839.","version":2}]}]}`, storeID)
		case base + "recalls":
			var in struct {
				StoreID int64
				Path    string
			}
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.StoreID != storeID || in.Path != "topic.md" {
				t.Errorf("recall %+v %v", in, err)
			}
			recalls.Add(1)
			w.WriteHeader(202)
		case base + "changes":
			if r.URL.Query().Get("storeId") != fmt.Sprint(storeID) || r.URL.Query().Get("afterRevision") != "2" {
				t.Error(r.URL)
			}
			fmt.Fprint(w, `{"currentRevision":3,"hasMore":false}`)
		case base + "mutations":
			var in aw.MemoryMutationRequest
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Operation != "UPDATE" || in.Path != "topic.md" || in.ExpectedVersion == nil || in.IdempotencyKey == "" || r.URL.Query().Get("storeId") != fmt.Sprint(storeID) {
				t.Errorf("mutation %+v %v", in, err)
				w.WriteHeader(400)
				return
			}
			if *in.ExpectedVersion != 2 {
				w.WriteHeader(409)
				return
			}
			if writes.Load() > 0 && (in.ContentMD != "verified" || in.IdempotencyKey != "native-verify") {
				t.Errorf("unexpected native mutation %+v", in)
			}
			writes.Add(1)
			fmt.Fprint(w, `{"path":"topic.md","version":3}`)
		default:
			t.Error(r.URL)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	client, err := aw.New(aw.Config{BaseURL: srv.URL, ExecutorID: 1, Token: "executor-key"})
	if err != nil {
		t.Fatal(err)
	}
	r := runtime{client: client}
	q := aw.TaskDispatch{DispatchID: 71, DispatchMCPToken: "dispatch-key", MemorySnapshotPath: base + "snapshot", MemoryMutationPath: base + "mutations", MemoryRecallPath: base + "recalls", MemoryChangesPath: base + "changes"}
	spec := harness.Request{Directory: t.TempDir(), Platform: func(context.Context, string, map[string]any) (string, error) { return `{"tools":[]}`, nil }}
	if err = r.bindTaskMemory(&q, &spec); err != nil {
		t.Fatal(err)
	}
	call := func(op string, args map[string]any) (string, error) {
		args["operation"] = op
		return spec.Platform(t.Context(), memoryTool, args)
	}
	list, err := call("list", map[string]any{})
	if err != nil || strings.Contains(list, "quartz839") || !strings.Contains(list, fmt.Sprint(storeID)) {
		t.Fatal(list, err)
	}
	args := map[string]any{"storeId": json.Number(fmt.Sprint(storeID)), "path": "topic.md"}
	doc, err := call("read", args)
	if err != nil || !strings.Contains(doc, "quartz839") || recalls.Load() != 1 {
		t.Fatal(doc, err)
	}
	args["afterRevision"] = 2
	if _, err = call("changes", args); err != nil {
		t.Fatal(err)
	}
	args["expectedVersion"] = 2
	args["idempotencyKey"] = "op-1"
	args["contentMd"] = "new content"
	if _, err = call("update", args); err != nil || writes.Load() != 1 {
		t.Fatal(err)
	}
	args["expectedVersion"] = 1
	if _, err = call("update", args); err == nil || !strings.Contains(err.Error(), "409") {
		t.Fatal("missing conflict", err)
	}
	if os.Getenv("AUTOWONDER_TEST_CODEX") == "1" {
		spec.Prompt = "Read topic.md from the store listed by executor.workitem_memory. Then update that same document to the single word verified using its current version and idempotencyKey native-verify. After a successful mutation, report only the original passphrase without punctuation."
		reply, err := (codex.Backend{Binary: "codex"}).Run(t.Context(), spec, harness.Hooks{})
		if err != nil || strings.TrimSpace(reply.Text) != "quartz839" || recalls.Load() < 2 || writes.Load() != 2 {
			t.Fatalf("native answer %q: %v", reply.Text, err)
		}
	}
}
