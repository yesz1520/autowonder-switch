package executor

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

func TestDispatchDebugLog(t *testing.T) {
	for _, channel := range []string{"DIRECT", "RELAY"} {
		t.Run(channel, func(t *testing.T) {
			q := aw.TaskDispatch{DispatchID: 7, AgentID: 3, DebugLog: json.RawMessage(`{"enabled":true,"maxBytes":700}`)}
			d := newDispatchLog(&q)
			for i := range 30 {
				d.event("model.message.bytes", int64(i))
			}
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			defer server.Close()
			var issued aw.DebugLogUploadRequest
			var uploaded []byte
			mux.HandleFunc("POST /api/daemon/dispatches/7/debug-log-upload", func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewDecoder(r.Body).Decode(&issued); err != nil {
					t.Error(err)
				}
				if channel == "RELAY" {
					w.WriteHeader(503)
					return
				}
				json.NewEncoder(w).Encode(aw.DebugLogUploadLink{UploadURL: server.URL + "/signed"})
			})
			mux.HandleFunc("PUT /signed", func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "" || r.URL.RawQuery != "" {
					t.Error("unexpected upload credentials")
				}
				uploaded, _ = io.ReadAll(r.Body)
			})
			mux.HandleFunc("POST /api/daemon/dispatches/7/artifacts", func(w http.ResponseWriter, r *http.Request) {
				f, _, err := r.FormFile("files")
				if err != nil {
					t.Error(err)
					return
				}
				defer f.Close()
				uploaded, _ = io.ReadAll(f)
				var meta []map[string]any
				err = json.Unmarshal([]byte(r.FormValue("filesMetadata")), &meta)
				if err != nil || len(meta) != 1 || meta[0]["sha256"] != issued.SHA256 || meta[0]["path"] != "debug/agent-3-7.log.gz" {
					t.Errorf("metadata %v: %v", meta, err)
				}
				fmt.Fprint(w, `{"files":[{"index":0,"status":"ACCEPTED"}]}`)
			})
			c, err := aw.New(aw.Config{ExecutorID: 1, Token: "test", BaseURL: server.URL})
			if err != nil {
				t.Fatal(err)
			}
			var out struct {
				Status, Channel, SHA256 string
				Truncated               bool
			}
			err = json.Unmarshal(d.upload(t.Context(), c, &q, &aw.TaskResult{Success: true, ResultID: aw.NewResultID()}), &out)
			if err != nil || out.Status != "UPLOADED" || out.Channel != channel || !out.Truncated || len(uploaded) != int(issued.SizeBytes) || issued.SHA256 != fmt.Sprintf("%x", sha256.Sum256(uploaded)) {
				t.Fatalf("metadata=%+v issued=%+v err=%v", out, issued, err)
			}
			gz, err := gzip.NewReader(bytes.NewReader(uploaded))
			if err != nil {
				t.Fatal(err)
			}
			defer gz.Close()
			txt, err := io.ReadAll(gz)
			if err != nil || len(txt) > 700 || !bytes.Contains(txt, []byte("model.message.bytes")) {
				t.Fatalf("log %q: %v", txt, err)
			}
		})
	}
}
