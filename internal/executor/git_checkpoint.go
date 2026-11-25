package executor

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const archiveLimit = 50 << 20

type checkpointMeta struct {
	Schema   string     `json:"schemaVersion"`
	Dispatch int64      `json:"dispatchId"`
	Attempt  int64      `json:"attempt"`
	Result   string     `json:"resultId"`
	Provider string     `json:"provider"`
	Session  string     `json:"providerSessionId"`
	Artifact string     `json:"artifact"`
	Repos    []revision `json:"repos,omitempty"`
}

func repoCheckpoint(ctx context.Context, id, attempt int64, result, session, answer string, repos []checkout) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	var total int64
	add := func(name string, data []byte) error {
		n := int64(len(data))
		total += n
		if total > 128<<20 {
			return errors.New("expanded checkpoint exceeds 128 MiB")
		}
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: n}); err != nil {
			return err
		}
		_, err := tw.Write(data)
		return err
	}
	meta := checkpointMeta{"autowonder.runtimeCheckpoint.v1", id, attempt, result, "codex", session, "result.md", nil}
	for _, c := range repos {
		meta.Repos = append(meta.Repos, c.revision())
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	if err = add("checkpoint.json", raw); err != nil {
		return nil, err
	}
	if err = add("result.md", []byte(answer)); err != nil {
		return nil, err
	}
	for i, c := range repos {
		prefix := fmt.Sprintf("repos/%d/", i)
		if c.head != c.base {
			// Name the exact local commit, even if no branch was pushed.
			if _, err = c.git(ctx, "update-ref", "refs/autowonder/checkpoint", c.head); err != nil {
				return nil, err
			}
			path := filepath.Join(c.gitDir, "delivery.bundle")
			if _, err = c.git(ctx, "bundle", "create", path, "refs/autowonder/checkpoint", "^"+c.base); err != nil {
				return nil, err
			}
			st, err := os.Stat(path)
			if err != nil {
				return nil, err
			}
			if st.Size() > archiveLimit {
				return nil, errors.New("Git bundle exceeds platform checkpoint limit")
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			if err = add(prefix+"head.bundle", b); err != nil {
				return nil, err
			}
		}
		c.base = c.head
		diff, err := c.patch(ctx)
		if err != nil {
			return nil, err
		}
		if err = add(prefix+"working.patch", diff); err != nil {
			return nil, err
		}
	}
	if err = tw.Close(); err != nil {
		return nil, err
	}
	if err = gz.Close(); err != nil {
		return nil, err
	}
	if buf.Len() > archiveLimit {
		return nil, errors.New("checkpoint exceeds platform 50 MiB limit")
	}
	return buf.Bytes(), nil
}
func readCheckpoint(path string, id int64) (checkpointMeta, map[string][]byte, error) {
	var meta checkpointMeta
	f, err := os.Open(path)
	if err != nil {
		return meta, nil, err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return meta, nil, err
	}
	defer gz.Close()
	rd := tar.NewReader(gz)
	entries := map[string][]byte{}
	var size int64
	for {
		h, e := rd.Next()
		if errors.Is(e, io.EOF) {
			break
		}
		if e != nil {
			return meta, nil, e
		}
		size += h.Size
		if size > 128<<20 || h.Size < 0 || len(entries) > 204 || h.Typeflag != tar.TypeReg || entries[h.Name] != nil {
			return meta, nil, errors.New("invalid checkpoint archive")
		}
		data, e := io.ReadAll(rd)
		if e != nil {
			return meta, nil, e
		}
		entries[h.Name] = data
	}
	if err = json.Unmarshal(entries["checkpoint.json"], &meta); err != nil {
		return meta, nil, err
	}
	if meta.Schema != "autowonder.runtimeCheckpoint.v1" || meta.Dispatch != id || meta.Artifact != "result.md" || len(entries["result.md"]) == 0 {
		return meta, nil, errors.New("checkpoint identity or answer mismatch")
	}
	return meta, entries, nil
}

// Reconstruct candidates away from the active workspace. A bad candidate can
// be discarded without leaking staged changes into the next restore attempt.
func restoreRepos(ctx context.Context, path string, id int64, repos []checkout) (string, error) {
	meta, entries, err := readCheckpoint(path, id)
	if err != nil {
		return "", err
	}
	if len(meta.Repos) != len(repos) {
		return "", errors.New("checkpoint repository set differs from dispatch")
	}
	if len(repos) == 0 {
		return string(entries["result.md"]), nil
	}
	scratch, err := os.MkdirTemp(filepath.Dir(repos[0].gitDir), ".restore-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(scratch)
	byName := map[string]int{}
	for i, c := range repos {
		byName[c.name] = i
	}
	rebuilt := make([]checkout, len(repos))
	for i, v := range meta.Repos {
		n, ok := byName[v.Name]
		if !ok || !oidPattern.MatchString(v.Base) || !oidPattern.MatchString(v.Head) {
			return "", errors.New("invalid checkpoint repository")
		}
		delete(byName, v.Name)
		c := repos[n]
		if c.base != v.Base && c.base != v.Head {
			return "", fmt.Errorf("%s checkpoint base differs from package revision", c.name)
		}
		gd := filepath.Join(scratch, fmt.Sprintf("git-%d", n))
		wd := filepath.Join(scratch, fmt.Sprintf("work-%d", n))
		if _, err = git(ctx, "clone", "--mirror", "--no-local", "--", c.gitDir, gd); err != nil {
			return "", err
		}
		c.gitDir = gd
		c.workDir = wd
		if err = os.Mkdir(wd, 0700); err != nil {
			return "", err
		}
		prefix := fmt.Sprintf("repos/%d/", i)
		patch, ok := entries[prefix+"working.patch"]
		if !ok {
			return "", errors.New("missing checkpoint patch")
		}
		if v.Head != v.Base {
			bundle := filepath.Join(gd, "restore.bundle")
			if err = os.WriteFile(bundle, entries[prefix+"head.bundle"], 0600); err != nil {
				return "", err
			}
			if _, err = c.git(ctx, "fetch", "--no-tags", "--", bundle, "refs/autowonder/checkpoint"); err != nil {
				return "", fmt.Errorf("fetch checkpoint bundle: %w", err)
			}
			out, e := c.git(ctx, "rev-parse", "FETCH_HEAD")
			if e != nil || strings.TrimSpace(string(out)) != v.Head {
				return "", errors.New("checkpoint bundle HEAD mismatch")
			}
		}
		if _, err = c.git(ctx, "merge-base", "--is-ancestor", v.Base, v.Head); err != nil {
			return "", errors.New("checkpoint head does not descend from base")
		}
		if _, err = c.git(ctx, "update-ref", "refs/autowonder/restored", v.Head); err != nil {
			return "", err
		}
		if _, err = c.git(ctx, "read-tree", v.Head); err != nil {
			return "", err
		}
		if _, err = c.git(ctx, "checkout-index", "--all", "--index"); err != nil {
			return "", err
		}
		if len(patch) > 0 {
			p := filepath.Join(gd, "restore.patch")
			if err = os.WriteFile(p, patch, 0600); err != nil {
				return "", err
			}
			if _, err = c.git(ctx, "apply", "--index", "--binary", "--", p); err != nil {
				return "", fmt.Errorf("apply checkpoint patch: %w", err)
			}
		}
		c.base = v.Base
		c.head = v.Head
		c.branch = v.Branch
		c.verified = ""
		rebuilt[n] = c
	}
	// Prepare Git's final paths before swapping directories. Roll back all
	// previously installed repositories if a later rename fails.
	type move struct{ from, to string }
	var moves []move
	for i, c := range rebuilt {
		old := repos[i]
		if _, err = c.git(ctx, "config", "core.bare", "false"); err != nil {
			return "", err
		}
		if _, err = c.git(ctx, "config", "core.worktree", old.workDir); err != nil {
			return "", err
		}
		if _, err = c.git(ctx, "update-ref", "--no-deref", "HEAD", c.head); err != nil {
			return "", err
		}
		if err = writeFile(filepath.Join(c.workDir, ".git"), []byte("gitdir: "+old.gitDir+"\n")); err != nil {
			return "", err
		}
		for _, pair := range [][2]string{{old.gitDir, c.gitDir}, {old.workDir, c.workDir}} {
			moves = append(moves, move{pair[0], pair[1] + ".old"}, move{pair[1], pair[0]})
		}
		c.gitDir = old.gitDir
		c.workDir = old.workDir
		rebuilt[i] = c
	}
	for i, m := range moves {
		if err = os.Rename(m.from, m.to); err != nil {
			for n := i - 1; n >= 0; n-- {
				undo := moves[n]
				err = errors.Join(err, os.Rename(undo.to, undo.from))
			}
			return "", err
		}
	}
	copy(repos, rebuilt)

	return string(entries["result.md"]), nil
}
