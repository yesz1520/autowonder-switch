package executor

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

type repository struct {
	Mode                                                                         string
	Name, URL, Ref, PinnedBaseCommit, DeliveryBranch, DeliveryUnpushedHeadCommit string
	AllowedBranchPatterns, SourceRevisionFallbackRefs                            []string
	AllowCommit, AllowPush, AllowNetwork                                         bool
}
type permissions struct{ AllowCommit, AllowPush, AllowNetwork *bool }
type checkout struct {
	name, gitDir, workDir, base, remote, head, branch, verified string
	spec                                                        repository
}

var oidPattern = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

func permitted(allowed bool, caps ...*bool) bool {
	for _, v := range caps {
		if v != nil && !*v {
			return false
		}
	}
	return allowed
}
func git(ctx context.Context, args ...string) ([]byte, error) {
	return gitCommand(ctx, false, false, args...)
}
func gitCommand(ctx context.Context, network, local bool, args ...string) ([]byte, error) {
	options := []string{"-c", "core.hooksPath=" + os.DevNull, "-c", "protocol.allow=never", "-c", "protocol.file.allow=never", "-c", "gc.auto=0"}
	if !network || local {
		options = append(options, "-c", "protocol.file.allow=always")
	}
	if network && !local {
		options = append(options, "-c", "protocol.http.allow=always", "-c", "protocol.https.allow=always", "-c", "protocol.ssh.allow=always")
	}
	cmd := exec.CommandContext(ctx, "git", append(options, args...)...)
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0", "GIT_LFS_SKIP_SMUDGE=1", "GIT_AUTHOR_NAME=AutoWonder", "GIT_COMMITTER_NAME=AutoWonder", "GIT_AUTHOR_EMAIL=executor@localhost", "GIT_COMMITTER_EMAIL=executor@localhost"}
	if network {
		for _, key := range []string{"HOME", "USERPROFILE", "SSH_AUTH_SOCK", "XDG_CONFIG_HOME", "HTTPS_PROXY", "NO_PROXY", "GIT_SSL_CAINFO"} {
			if v := os.Getenv(key); v != "" {
				cmd.Env = append(cmd.Env, key+"="+v)
			}
		}
	} else {
		cmd.Env = append(cmd.Env, "GIT_CONFIG_GLOBAL="+os.DevNull)
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git operation failed: %w", err)
	}
	return out, nil
}
func validRemote(raw string) bool {
	if strings.ContainsAny(raw, "\r\n\t ") || strings.HasPrefix(raw, "-") {
		return false
	}
	if scpRemote.MatchString(raw) {
		return true
	}
	u, e := url.Parse(raw)
	if e != nil {
		return false
	}
	if u.User != nil {
		if _, ok := u.User.Password(); ok || strings.IndexFunc(u.User.Username(), unicode.IsControl) >= 0 {
			return false
		}
	}
	return (u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "ssh") && u.Hostname() != "" && !strings.HasPrefix(u.Hostname(), "-") && u.Path != "" && u.RawQuery == "" && u.Fragment == "" && (u.User == nil || u.Scheme == "ssh" && !strings.HasPrefix(u.User.Username(), "-"))
}

var scpRemote = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.-]*@[a-zA-Z0-9][a-zA-Z0-9.-]*:[^:]+$`)

func validName(s string) bool {
	return safeFile(s) && strings.TrimSpace(s) == s && !strings.HasPrefix(s, "~") && !strings.HasPrefix(s, "%") && !strings.EqualFold(s, ".git") && !strings.Contains(s, ":") && strings.IndexFunc(s, unicode.IsControl) < 0
}

// Resolve the package revision in an executor-owned object store.
// Only the working files are writable by a model turn.
func prepareRepos(ctx context.Context, pkg, dir, work string, sources map[string]string, frame ...*aw.TaskDispatch) ([]checkout, error) {
	z, err := zip.OpenReader(pkg)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	var spec struct{ Repos []repository }
	var policy permissions
	for _, v := range z.File {
		if v.Name == "policy.json" {
			rd, e := v.Open()
			if e != nil {
				return nil, e
			}
			e = json.NewDecoder(io.LimitReader(rd, 1<<20)).Decode(&policy)
			rd.Close()
			if e != nil {
				return nil, e
			}
		}
		if v.Name == "repos.json" {
			rd, err := v.Open()
			if err != nil {
				return nil, err
			}
			err = json.NewDecoder(io.LimitReader(rd, 2<<20)).Decode(&spec)
			rd.Close()
			if err != nil {
				return nil, err
			}
		}
	}
	return prepareRepoList(ctx, spec.Repos, dir, work, sources, policy, frame...)
}
func prepareRepoList(ctx context.Context, list []repository, dir, work string, sources map[string]string, policy permissions, frame ...*aw.TaskDispatch) ([]checkout, error) {
	var err error
	var result []checkout
	seen := map[string]bool{}
	for _, v := range list {
		if !validName(v.Name) || seen[strings.ToLower(v.Name)] {
			return nil, errors.New("invalid or duplicate repository name")
		}
		seen[strings.ToLower(v.Name)] = true
		var envelope permissions
		if len(frame) > 0 {
			q := frame[0]
			envelope = permissions{q.AllowCommit, q.AllowPush, q.AllowNetwork}
		}
		v.AllowCommit = permitted(v.AllowCommit, policy.AllowCommit, envelope.AllowCommit)
		v.AllowPush = permitted(v.AllowPush, policy.AllowPush, envelope.AllowPush)
		v.AllowNetwork = permitted(v.AllowNetwork, policy.AllowNetwork, envelope.AllowNetwork)
		source, mapped := sources[v.Name]
		gd := filepath.Join(dir, "git-"+v.Name)
		wd := filepath.Join(work, v.Name)
		remote := v.URL
		if mapped {
			if !filepath.IsAbs(source) {
				return nil, errors.New("local repository path must be absolute")
			}
			if _, err = git(ctx, "clone", "--bare", "--no-local", "--", source, gd); err != nil {
				return nil, err
			}
			if remote == "" {
				o, e := git(ctx, "-C", source, "remote", "get-url", "origin")
				if e == nil {
					remote = strings.TrimSpace(string(o))
				} else if bare, e := git(ctx, "-C", source, "rev-parse", "--is-bare-repository"); e == nil && strings.TrimSpace(string(bare)) == "true" {
					remote = source
				}
			}
		} else {
			if !v.AllowNetwork || !validRemote(remote) {
				return nil, fmt.Errorf("repo %q requires a local mapping or an allowed HTTP(S)/SSH source", v.Name)
			}
			if _, err = gitCommand(ctx, true, false, "clone", "--bare", "--", remote, gd); err != nil {
				return nil, fmt.Errorf("clone %s: %w", v.Name, err)
			}
		}
		// A file transport can only come from operator-owned local configuration.
		trustedFile := mapped && v.URL == "" && filepath.IsAbs(remote)
		if remote != "" && !trustedFile && !validRemote(remote) {
			return nil, errors.New("unsupported repository remote")
		}
		ref := v.Ref
		if ref == "" {
			ref = "HEAD"
		}
		refs := append([]string{ref}, v.SourceRevisionFallbackRefs...)
		if v.PinnedBaseCommit != "" {
			if !oidPattern.MatchString(v.PinnedBaseCommit) {
				return nil, errors.New("invalid pinned commit")
			}
			refs = []string{v.PinnedBaseCommit}
		}
		var oid []byte
		for _, candidate := range refs {
			if candidate == "" {
				candidate = "HEAD"
			}
			if candidate != "HEAD" && !oidPattern.MatchString(candidate) {
				if strings.HasPrefix(candidate, "-") {
					return nil, errors.New("invalid checkout ref")
				}
				if _, e := git(ctx, "check-ref-format", "refs/heads/"+candidate); e != nil {
					return nil, e
				}
			}
			needFresh := mapped && v.AllowNetwork && remote != "" && !oidPattern.MatchString(candidate)
			if !needFresh {
				oid, err = git(ctx, "--git-dir", gd, "rev-parse", "--verify", "--end-of-options", candidate+"^{commit}")
				if err == nil {
					break
				}
			}
			if v.AllowNetwork && remote != "" {
				_, err = gitCommand(ctx, true, trustedFile, "--git-dir", gd, "fetch", "--no-tags", "--", remote, candidate)
				if err == nil {
					oid, err = git(ctx, "--git-dir", gd, "rev-parse", "--verify", "FETCH_HEAD^{commit}")
					if err == nil {
						break
					}
				}
			}

		}
		if err != nil {
			return nil, fmt.Errorf("requested revision for %q unavailable", v.Name)
		}
		base := strings.TrimSpace(string(oid))
		co := checkout{name: v.Name, gitDir: gd, workDir: wd, base: base, head: base, remote: remote, spec: v}
		if err = os.MkdirAll(wd, 0700); err != nil {
			return nil, err
		}
		if _, err = co.git(ctx, "read-tree", co.base); err != nil {
			return nil, err
		}
		if _, err = co.git(ctx, "checkout-index", "--all", "--prefix="+wd+string(os.PathSeparator)); err != nil {
			return nil, err
		}
		if err = co.attach(ctx); err != nil {
			return nil, err
		}
		result = append(result, co)
	}
	return result, nil
}
func (c checkout) git(ctx context.Context, args ...string) ([]byte, error) {
	return git(ctx, append([]string{"-C", c.workDir, "--git-dir", c.gitDir, "--work-tree", c.workDir}, args...)...)
}
func (c checkout) patch(ctx context.Context) ([]byte, error) {
	info, err := os.Lstat(c.workDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("repository working directory was replaced")
	}
	if _, err := c.git(ctx, "add", "-A", "--", "."); err != nil {
		return nil, err
	}
	p, err := c.git(ctx, "diff", "--cached", "--no-ext-diff", "--no-textconv", "--binary", c.base, "--")
	if err != nil {
		return nil, err
	}
	if len(p) > archiveLimit {
		return nil, errors.New("repository patch exceeds 50 MiB")
	}
	return p, nil
}

// Keep Git metadata outside the writable sandbox while allowing normal Git
// inspection from the working directory, including at a pinned detached HEAD.
func (c checkout) attach(ctx context.Context) error {
	if _, err := c.git(ctx, "update-ref", "--no-deref", "HEAD", c.head); err != nil {
		return err
	}
	if _, err := c.git(ctx, "update-ref", "refs/autowonder/base", c.base); err != nil {
		return err
	}
	if _, err := c.git(ctx, "config", "core.bare", "false"); err != nil {
		return err
	}
	if _, err := c.git(ctx, "config", "core.worktree", c.workDir); err != nil {
		return err
	}
	return writeFile(filepath.Join(c.workDir, ".git"), []byte("gitdir: "+c.gitDir+"\n"))
}
