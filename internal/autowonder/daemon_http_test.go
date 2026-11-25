package autowonder

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// Check HTTP method, credential scope and content independently of the client
// against the routes in the 1.3.0 Daemon*Controller classes.
func TestDaemonRoutes(t *testing.T) {
	ctx := t.Context()
	type operation struct {
		method, path, auth, bodyKey string
		call                        func() error
	}
	var c *Client
	var m *Memory
	cases := []operation{
		{"GET", "/executors/42/environment-variables", "executor", "", func() error { _, e := c.Environment(ctx); return e }},
		{"POST", "/tasks/71/usage", "executor", "usage", func() error { return c.ReportUsage(ctx, 71, []UsageEntry{{Model: "test"}}) }},
		{"GET", "workitem-identity", "query", "", func() error { _, e := c.WorkitemIdentity(ctx, 71); return e }},
		{"GET", "task-data?namespace=n&key=k", "query", "", func() error { _, e := c.GetTaskData(ctx, 71, "n", "k"); return e }},
		{"GET", "task-data?namespace=n&afterId=2&limit=3", "query", "", func() error { _, e := c.ListTaskData(ctx, 71, "n", 2, 3); return e }},
		{"PUT", "task-data", "query", "expectedVersion", func() error {
			v := 0
			_, e := c.PutTaskData(ctx, 71, TaskDataWrite{Namespace: "n", Key: "k", ExpectedVersion: &v})
			return e
		}},
		{"POST", "comments", "query", "contentMd", func() error { _, e := c.AddComment(ctx, 71, Comment{ContentMD: "ok"}); return e }},
		{"POST", "workitem-status", "query", "status", func() error { _, e := c.SetWorkitemStatus(ctx, 71, "x"); return e }},
		{"POST", "acceptance", "query", "[]", func() error { _, e := c.ReportAcceptance(ctx, 71, []AcceptanceReport{{RequirementKey: "r"}}); return e }},
		{"POST", "requests", "executor", "message", func() error { _, e := c.CreateHumanRequest(ctx, 71, HumanRequestInput{Message: "?"}); return e }},
		{"GET", "requests/4", "executor", "", func() error { _, e := c.GetHumanRequest(ctx, 71, 4); return e }},
		{"POST", "requests/4/cancel", "executor", "", func() error { return c.CancelHumanRequest(ctx, 71, 4) }},
		{"POST", "debug-log-upload", "query", "sizeBytes", func() error { _, e := c.DebugLogUpload(ctx, 71, DebugLogUploadRequest{SizeBytes: 15}); return e }},
		{"OPTIONS", "artifacts", "query", "", func() error {
			v, e := c.LogArchiveCapability(ctx, 71)
			if e == nil && (!v.Supported || v.MaxDecodedBytes != 128) {
				t.Error("archive headers lost")
			}
			return e
		}},
		{"GET", "memory/snapshot", "dispatch", "", func() error { _, e := m.Snapshot(ctx); return e }},
		{"GET", "memory/changes?storeId=9&afterRevision=10&limit=4", "dispatch", "", func() error { _, e := m.Changes(ctx, 9, 10, 4); return e }},
		{"POST", "memory/mutations?storeId=9", "dispatch", "operation", func() error { _, e := m.Mutate(ctx, 9, MemoryMutationRequest{Operation: "UPDATE"}); return e }},
		{"POST", "memory/recalls", "dispatch", "path", func() error { return m.Recall(ctx, 9, "topic") }},
		{"POST", "memory/maintenance/claims", "dispatch", "storeId", func() error { _, e := m.ClaimMaintenance(ctx, MemoryMaintenanceLeaseRequest{StoreID: 9}); return e }},
		{"POST", "memory/maintenance/releases", "dispatch", "leaseId", func() error { return m.ReleaseMaintenance(ctx, MemoryMaintenanceLeaseRequest{LeaseID: "abc"}) }},
		{"POST", "memory/maintenance/renewals", "dispatch", "leaseId", func() error { return m.RenewMaintenance(ctx, MemoryMaintenanceLeaseRequest{LeaseID: "abc"}) }},
		{"POST", "memory/imports/manifest", "dispatch", "sanitizedContentSha256", func() error { _, e := m.ImportManifest(ctx, MemoryImportManifestRequest{}); return e }},
		{"POST", "memory/imports/snapshots", "dispatch", "sanitizedContent", func() error { return m.ImportSnapshot(ctx, MemoryImportSnapshotRequest{}) }},
		{"POST", "memory/imports/completions", "dispatch", "redactionCount", func() error { return m.CompleteImport(ctx, MemoryImportCompletionRequest{}) }},
		{"POST", "memory/imports/claims", "dispatch", "sourceId", func() error { _, e := m.ClaimImport(ctx, MemoryImportClaimRequest{}); return e }},
		{"POST", "memory/imports/renewals", "dispatch", "sourceId", func() error { return m.RenewImport(ctx, MemoryImportClaimRequest{}) }},
		{"POST", "memory/imports/releases", "dispatch", "sourceId", func() error { return m.ReleaseImport(ctx, MemoryImportClaimRequest{}) }},
	}
	for _, op := range cases {
		t.Run(op.method+" "+op.path, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				prefix := "/api/daemon/dispatches/71/"
				if op.path[0] == '/' {
					prefix = "/api/daemon"
				}
				want, _ := url.Parse(prefix + op.path)
				q := want.Query()
				auth := "Bearer " + op.auth
				if op.auth == "query" {
					auth = ""
					q.Set("token", "executor")
				}
				if r.Method != op.method || r.URL.Path != want.Path || r.URL.Query().Encode() != q.Encode() || r.Header.Get("Authorization") != auth {
					t.Errorf("wrong route or auth: %s %s", r.Method, r.URL.Path)
				}
				if op.bodyKey != "" {
					if r.Header.Get("Content-Type") != "application/json" {
						t.Error("missing JSON content type")
					}
					data, err := io.ReadAll(r.Body)
					if err != nil {
						t.Error(err)
					}
					if op.bodyKey == "[]" {
						var a []AcceptanceReport
						if json.Unmarshal(data, &a) != nil || len(a) != 1 {
							t.Error("wrong array")
						}
					} else {
						var fields map[string]json.RawMessage
						if json.Unmarshal(data, &fields) != nil || fields[op.bodyKey] == nil {
							t.Errorf("missing %s", op.bodyKey)
						}
					}
				}
				w.Header().Set("X-AutoWonder-Log-Archive", "1")
				w.Header().Set("X-AutoWonder-Log-Archive-Max-Decoded-Bytes", "128")
				switch op.path {
				case "memory/recalls", "memory/maintenance/releases", "memory/maintenance/renewals", "memory/imports/snapshots", "memory/imports/completions", "memory/imports/releases", "memory/imports/renewals":
					w.WriteHeader(http.StatusAccepted)
				default:
					io.WriteString(w, `{"status":"CANCELED"}`)
				}
			}))
			defer srv.Close()
			var err error
			c, err = New(Config{BaseURL: srv.URL, ExecutorID: 42, Token: "executor"})
			if err != nil {
				t.Fatal(err)
			}
			m, err = c.Memory(71, "dispatch")
			if err != nil {
				t.Fatal(err)
			}
			if err := op.call(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
