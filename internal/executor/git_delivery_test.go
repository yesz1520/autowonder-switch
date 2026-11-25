package executor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/harness/codex"
)

func testGit(t *testing.T, args ...string) string {
	t.Helper()
	out, e := git(t.Context(), args...)
	if e != nil {
		t.Fatalf("git %v: %v", args, e)
	}
	return strings.TrimSpace(string(out))
}

type gitFixture struct {
	url, remote, base, pkg string
	spec                   repository
}

func newGitFixture(t *testing.T) gitFixture {
	t.Helper()
	root := t.TempDir()
	remote := filepath.Join(root, "origin.git")
	seed := filepath.Join(root, "seed")
	testGit(t, "init", "--bare", "-b", "main", remote)
	testGit(t, "init", "-b", "main", seed)
	if err := os.WriteFile(filepath.Join(seed, "README.md"), []byte("baseline\n"), 0600); err != nil {
		t.Fatal(err)
	}
	testGit(t, "-C", seed, "add", ".")
	testGit(t, "-C", seed, "commit", "-m", "initial")
	base := testGit(t, "-C", seed, "rev-parse", "HEAD")
	testGit(t, "-C", seed, "push", remote, "main")
	testGit(t, "--git-dir", remote, "config", "http.receivepack", "true")
	backend := filepath.Join(testGit(t, "--exec-path"), "git-http-backend")
	server := httptest.NewTLSServer(&cgi.Handler{Path: backend, Env: []string{"GIT_PROJECT_ROOT=" + root, "GIT_HTTP_EXPORT_ALL=1"}})
	t.Cleanup(server.Close)
	cert := filepath.Join(root, "ca.pem")
	if err := os.WriteFile(cert, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_SSL_CAINFO", cert)
	spec := repository{Name: "project", URL: server.URL + "/origin.git", Ref: "main", AllowNetwork: true, AllowCommit: true, AllowPush: true, AllowedBranchPatterns: []string{"feature/*"}}
	raw, _ := json.Marshal(map[string]any{"repos": []repository{spec}})
	p := filepath.Join(root, "task.zip")
	if err := os.WriteFile(p, archive(t, map[string]string{"repos.json": string(raw), "workitem.md": "Create violet.txt containing just violet and push the change to feature/test."}), 0600); err != nil {
		t.Fatal(err)
	}
	return gitFixture{server.URL, remote, base, p, spec}
}
func prepareFixture(t *testing.T, p string) (string, []checkout) {
	t.Helper()
	dir := t.TempDir()
	repos, e := prepareRepos(t.Context(), p, dir, filepath.Join(dir, "workspace"), nil)
	if e != nil {
		t.Fatal(e)
	}
	return dir, repos
}
func TestGitDeliveryAndRecovery(t *testing.T) {
	f := newGitFixture(t)
	dir, repos := prepareFixture(t, f.pkg)
	if repos[0].base != f.base {
		t.Fatal("incorrect remote checkout")
	}
	c := &repos[0]
	newPath := filepath.Join(c.workDir, "binary")
	expected := []byte{0, 255, 4}
	os.WriteFile(newPath, expected, 0600)
	prepare := func(mode string) (preparedDelivery, error) {
		b, _ := json.Marshal(codeReply{Success: true, Summary: "edited", Repositories: []deliveryIntent{{"project", mode, "feature/test", "Change file"}}})
		return prepareDelivery(t.Context(), dir, "id", "thread", string(b), repos, nil)
	}
	_, err := prepare("patch")
	if err != nil || c.head != f.base {
		t.Fatalf("patch: %v", err)
	}
	i := deliveryIntent{"project", "push", "main", "test"}
	if c.validate(t.Context(), i) == nil {
		t.Fatal("branch policy bypass")
	}
	i.Branch = "feature/test"
	c.spec.AllowPush = false
	if c.validate(t.Context(), i) == nil {
		t.Fatal("push denied by repo")
	}
	c.spec.AllowPush = true
	d, err := prepare("commit")
	if err != nil {
		t.Fatal(err)
	}
	head := d.Repos[0].Head
	if head == f.base {
		t.Fatal("no commit")
	}
	if _, err := git(t.Context(), "--git-dir", f.remote, "show-ref", "--verify", "refs/heads/feature/test"); err == nil {
		t.Fatal("commit mode pushed")
	}
	os.WriteFile(filepath.Join(c.workDir, "untracked"), []byte("pending"), 0600)
	os.Remove(filepath.Join(c.workDir, "README.md"))
	raw, err := repoCheckpoint(t.Context(), "codex", 17, 1, "", "native", "paused", repos)
	if err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(dir, "pause.tar.gz")
	os.WriteFile(archivePath, raw, 0600)
	_, restored := prepareFixture(t, f.pkg)
	if _, err = restoreRepos(t.Context(), archivePath, 17, restored); err != nil {
		t.Fatal(err)
	}
	rc := restored[0]
	if rc.head != head {
		t.Fatal("local-only commit lost")
	}
	b, _ := os.ReadFile(filepath.Join(rc.workDir, "binary"))
	if !bytes.Equal(b, expected) {
		t.Fatal("binary changed")
	}
	b, _ = os.ReadFile(filepath.Join(rc.workDir, "untracked"))
	if string(b) != "pending" {
		t.Fatal("untracked file lost")
	}
	if _, err = os.Stat(filepath.Join(rc.workDir, "README.md")); !os.IsNotExist(err) {
		t.Fatal("deletion lost")
	}
	if err = c.publish(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err = c.publish(t.Context()); err != nil {
		t.Fatalf("retry: %v", err)
	}
	if testGit(t, "--git-dir", f.remote, "rev-parse", "refs/heads/feature/test") != head {
		t.Fatal("incorrect remote SHA")
	}
	// Advance the branch independently. Retrying an older delivery must never rewind it.
	tree := testGit(t, "--git-dir", f.remote, "rev-parse", head+"^{tree}")
	newer := testGit(t, "--git-dir", f.remote, "commit-tree", tree, "-p", head, "-m", "concurrent")
	testGit(t, "--git-dir", f.remote, "update-ref", "refs/heads/feature/test", newer)
	if err = c.publish(t.Context()); err == nil {
		t.Fatal("accepted non-fast-forward push")
	}
	if testGit(t, "--git-dir", f.remote, "rev-parse", "feature/test") != newer {
		t.Fatal("overwrote newer delivery")
	}
}

type codeStub struct{ stub }

func (*codeStub) Run(_ context.Context, q harness.Request, _ harness.Hooks) (harness.Reply, error) {
	if q.OutputSchema == nil {
		return harness.Reply{}, fmt.Errorf("missing delivery schema")
	}
	if e := os.WriteFile(filepath.Join(q.Directory, "project", "violet.txt"), []byte("violet\n"), 0600); e != nil {
		return harness.Reply{}, e
	}
	b, _ := json.Marshal(codeReply{true, "edited", []deliveryIntent{{"project", "push", "feature/test", "Add violet"}}})
	return harness.Reply{Text: string(b), Session: "session", Model: "model"}, nil
}
func TestExecuteCodeDispatch(t *testing.T) {
	f := newGitFixture(t)
	pkg, err := os.ReadFile(f.pkg)
	if err != nil {
		t.Fatal(err)
	}
	var revisions []revision
	mux := http.NewServeMux()
	mux.HandleFunc("/package", func(w http.ResponseWriter, _ *http.Request) { w.Write(pkg) })
	mux.HandleFunc("/api/daemon/dispatches/1/artifacts", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(2 << 20); err != nil {
			t.Error(err)
			return
		}
		var meta []struct{ Path string }
		json.Unmarshal([]byte(r.FormValue("filesMetadata")), &meta)
		var receipts []aw.FileReceipt
		for i, p := range r.MultipartForm.File["files"] {
			if meta[i].Path == "deliverables/runtime-source-revision.json" {
				file, e := p.Open()
				if e != nil {
					t.Error(e)
					return
				}
				var doc struct{ Repositories []revision }
				e = json.NewDecoder(file).Decode(&doc)
				file.Close()
				if e != nil {
					t.Error(e)
				}
				revisions = doc.Repositories
			}
			receipts = append(receipts, aw.FileReceipt{Index: i, Status: "ACCEPTED"})
		}
		json.NewEncoder(w).Encode(aw.ArtifactsReceipt{Files: receipts})
	})
	mux.HandleFunc("/api/daemon/dispatches/1/checkpoint", func(w http.ResponseWriter, r *http.Request) {
		f, _, err := r.FormFile("checkpoint")
		if err != nil {
			t.Error(err)
			return
		}
		defer f.Close()
		b, _ := io.ReadAll(f)
		json.NewEncoder(w).Encode(aw.CheckpointReceipt{CheckpointSeq: 1, SizeBytes: int64(len(b)), SHA256: fmt.Sprintf("sha256:%x", sha256.Sum256(b)), ResultID: r.FormValue("resultId"), Purpose: "COMPLETION"})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	client, _ := aw.New(aw.Config{BaseURL: srv.URL, ExecutorID: 1, Token: "demo"})
	var backend harness.Backend = &codeStub{}
	if os.Getenv("AUTOWONDER_TEST_CODEX") == "1" {
		path, e := exec.LookPath("codex")
		if e != nil {
			t.Fatal(e)
		}
		backend = codex.Backend{Binary: path}
	}
	dir := t.TempDir()
	ctx, stop := context.WithTimeout(t.Context(), 2*time.Minute)
	defer stop()
	r := &runtime{client: client, backend: backend, dir: dir}
	r.Endpoint, _ = aw.NewEndpoint(func() aw.Heartbeat { return aw.Heartbeat{} }, nil)
	q := &aw.TaskDispatch{DispatchID: 1, Attempt: 1, PackageID: "p", DownloadURL: srv.URL + "/package", ChecksumAlgorithm: "sha256", ChecksumScope: "zip_archive", Checksum: fmt.Sprintf("sha256:%x", sha256.Sum256(pkg))}
	result := &aw.TaskResult{ResultID: aw.NewResultID()}
	if err = r.execute(&job{ctx: ctx}, q, dir, result); err != nil {
		t.Fatal(err)
	}
	remoteHead := testGit(t, "--git-dir", f.remote, "rev-parse", "feature/test")
	if len(revisions) != 1 || revisions[0].RemoteHead != remoteHead || !strings.Contains(result.ResultSummary, remoteHead) || result.CheckpointSHA256 == "" {
		t.Fatalf("unverified result: %+v", result)
	}
	if got := testGit(t, "--git-dir", f.remote, "show", "feature/test:violet.txt"); got != "violet" {
		t.Fatal(got)
	}
}

func TestCheckoutPolicyAndMissingRevision(t *testing.T) {
	f := newGitFixture(t)
	local := filepath.Join(t.TempDir(), "checkout")
	testGit(t, "clone", f.remote, local)
	old := testGit(t, "-C", local, "rev-parse", "HEAD")
	tree := testGit(t, "--git-dir", f.remote, "rev-parse", old+"^{tree}")
	next := testGit(t, "--git-dir", f.remote, "commit-tree", tree, "-p", old, "-m", "new revision")
	testGit(t, "--git-dir", f.remote, "update-ref", "refs/heads/main", next)
	newPackage := func(v repository, policy string) string {
		t.Helper()
		b, _ := json.Marshal(map[string]any{"repos": []repository{v}})
		path := filepath.Join(t.TempDir(), "repos.zip")
		err := os.WriteFile(path, archive(t, map[string]string{"repos.json": string(b), "policy.json": policy}), 0600)
		if err != nil {
			t.Fatal(err)
		}
		return path
	}
	run := func(p string, q *aw.TaskDispatch) ([]checkout, error) {
		dir := t.TempDir()
		return prepareRepos(t.Context(), p, dir, filepath.Join(dir, "work"), map[string]string{"project": local}, q)
	}
	no := false
	v := f.spec
	v.Ref = next
	repos, err := run(newPackage(v, `{"allowPush":false}`), &aw.TaskDispatch{AllowCommit: &no})
	if err != nil {
		t.Fatal(err)
	}
	if repos[0].base != next || repos[0].spec.AllowPush || repos[0].spec.AllowCommit {
		t.Fatal("fetch or permission intersection failed")
	}
	if _, err = run(newPackage(v, "{}"), &aw.TaskDispatch{AllowNetwork: &no}); err == nil {
		t.Fatal("fetched missing pinned revision despite network denial")
	}
	v.PinnedBaseCommit = strings.Repeat("0", 40)
	v.SourceRevisionFallbackRefs = []string{"main"}
	if _, err = run(newPackage(v, "{}"), &aw.TaskDispatch{}); err == nil {
		t.Fatal("substituted baseline for a pinned revision")
	}
}

func TestDispatchHookStopsBeforeModel(t *testing.T) {
	const script = "echo hook-denied >&2; exit 1"
	const descriptor = "schemaVersion: autowonder.hook.v1\nname: gate\ntrigger: beforeAgentStart\ninterpreter: sh\ncommand: check.sh\n"
	identity := fmt.Sprintf("check.sh\x00sha256:%x\nhook.yaml\x00sha256:%x\n", sha256.Sum256([]byte(script)), sha256.Sum256([]byte(descriptor)))
	files := map[string]string{
		"workitem.md":                       "try to start",
		"hooks.json":                        fmt.Sprintf(`{"schemaVersion":"autowonder.hooks.v1","hooks":[{"name":"gate","enabled":true,"sha256":"sha256:%x"}]}`, sha256.Sum256([]byte(identity))),
		"capabilities/hooks/gate/check.sh":  script,
		"capabilities/hooks/gate/hook.yaml": descriptor,
	}
	pkg := archive(t, files)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.Write(pkg) }))
	defer server.Close()
	c, err := aw.New(aw.Config{BaseURL: server.URL, ExecutorID: 1, Token: "test"})
	if err != nil {
		t.Fatal(err)
	}
	backend := &stub{}
	r := runtime{client: c, backend: backend, dir: t.TempDir()}
	r.Endpoint, _ = aw.NewEndpoint(func() aw.Heartbeat { return aw.Heartbeat{} }, nil)
	q := aw.TaskDispatch{DispatchID: 1, Attempt: 1, PackageID: "p", DownloadURL: server.URL, ChecksumAlgorithm: "sha256", ChecksumScope: "zip_archive", Checksum: fmt.Sprintf("sha256:%x", sha256.Sum256(pkg))}
	err = r.execute(&job{ctx: t.Context()}, &q, r.dir, &aw.TaskResult{})
	if err == nil || !strings.Contains(err.Error(), "beforeAgentStart hook gate") || backend.calls.Load() != 0 {
		t.Fatalf("calls %d error %v", backend.calls.Load(), err)
	}
	data, err := os.ReadFile(filepath.Join(r.dir, "hooks-1", "logs", "beforeAgentStart", "gate", "stderr"))
	if err != nil || !strings.Contains(string(data), "hook-denied") {
		t.Fatalf("hook stderr %q: %v", data, err)
	}
}
