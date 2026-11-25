package autowonder

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDispatchHTTP(t *testing.T) {
	ctx := t.Context()
	archive := "checkpoint bytes"
	digest := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(archive)))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/signed" {
			if r.URL.Query().Has("token") || r.Header.Get("Authorization") != "" {
				t.Error("platform credential sent to storage")
			}
			io.WriteString(w, archive)
			return
		}
		if r.Method != "POST" || !strings.HasPrefix(r.URL.Path, "/api/daemon/dispatches/701/") || r.URL.Query().Get("token") != "secret" {
			http.Error(w, "bad request", 400)
			return
		}
		switch filepath.Base(r.URL.Path) {
		case "package-url":
			io.WriteString(w, `{"downloadUrl":"https://storage.example/task.zip","expiresInSeconds":900}`)
		case "checkpoint-url":
			if r.URL.Query().Get("checkpointSeq") != "3" {
				t.Error("missing sequence")
			}
			fmt.Fprintf(w, `{"checkpointSeq":3,"downloadUrl":"https://storage.example/cp","sha256":%q}`, digest)
		case "recovery-claim":
			io.WriteString(w, `{"allowed":true,"status":"RUNNING","environmentVariables":{"LANG":"C"}}`)
		case "checkpoint", "artifacts":
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Error(err)
				return
			}
			defer r.MultipartForm.RemoveAll()
			if filepath.Base(r.URL.Path) == "checkpoint" {
				f, _, err := r.FormFile("checkpoint")
				if err != nil {
					t.Error(err)
					return
				}
				defer f.Close()
				bytes, err := io.ReadAll(f)
				if err != nil || string(bytes) != archive || r.FormValue("purpose") != "COMPLETION" || r.FormValue("checkpointSeq") != "3" {
					t.Error("bad checkpoint upload")
				}
				fmt.Fprintf(w, `{"checkpointSeq":3,"sha256":%q,"sizeBytes":16,"purpose":"COMPLETION","resultId":%q}`, digest, r.FormValue("resultId"))
			} else {
				if len(r.MultipartForm.File["files"]) != 2 || r.FormValue("idempotencyKey") != "retry-1" || !json.Valid([]byte(r.FormValue("filesMetadata"))) {
					t.Error("bad artifact parts")
				}
				io.WriteString(w, `{"files":[{"index":0,"status":"ACCEPTED"},{"index":1,"status":"REJECTED","code":"INVALID_PATH"}]}`)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c, err := New(Config{BaseURL: srv.URL, ExecutorID: 42, Token: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	pkg, err := c.RefreshPackage(ctx, 701)
	if err != nil || pkg.ExpiresInSeconds != 900 {
		t.Fatal(pkg, err)
	}
	cp, err := c.RefreshCheckpoint(ctx, 701, 3)
	if err != nil || cp.SHA256 != digest {
		t.Fatal(cp, err)
	}
	claim, err := c.Claim(ctx, 701)
	if err != nil || !claim.Allowed || claim.EnvironmentVariables["LANG"] != "C" {
		t.Fatal(claim, err)
	}
	const resultID = "86ebc809-2d32-44e4-b456-29b870f878bd"
	receipt, err := c.UploadCheckpoint(ctx, 701, CheckpointUpload{File: UploadFile{"cp", strings.NewReader(archive)}, Seq: 3, Purpose: "COMPLETION", ResultID: resultID})
	if err != nil || receipt.ResultID != resultID {
		t.Fatal(receipt, err)
	}
	res, err := c.UploadArtifacts(ctx, 701, []UploadFile{{"first", strings.NewReader("a")}, {"second", strings.NewReader("b")}}, "retry-1", []json.RawMessage{json.RawMessage(`{"path":"first"}`), json.RawMessage(`{"path":"second"}`)})
	if err == nil || len(res.Files) != 2 || res.Files[1].Code != "INVALID_PATH" {
		t.Fatal(res, err)
	}
	dest := filepath.Join(t.TempDir(), "saved")
	if err = c.DownloadVerified(ctx, srv.URL+"/signed", digest, dest); err != nil {
		t.Fatal(err)
	}
	if err = c.DownloadVerified(ctx, srv.URL+"/signed", strings.Repeat("0", 64), dest); err == nil {
		t.Fatal("accepted bad hash")
	}
	data, err := os.ReadFile(dest)
	if err != nil || string(data) != archive {
		t.Fatal("lost verified file", err)
	}
	_, err = c.Claim(ctx, 702)
	var status *HTTPError
	if !errors.As(err, &status) || status.StatusCode != 400 {
		t.Fatal(err)
	}
}
func TestHTTPRedirectDoesNotForwardToken(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("followed authenticated redirect") }))
	defer target.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+r.URL.RequestURI(), http.StatusTemporaryRedirect)
	}))
	defer origin.Close()
	c, err := New(Config{BaseURL: origin.URL, ExecutorID: 1, Token: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Claim(context.Background(), 1)
	var status *HTTPError
	if !errors.As(err, &status) || status.StatusCode != 307 || strings.Contains(err.Error(), "secret") {
		t.Fatal(err)
	}
}
