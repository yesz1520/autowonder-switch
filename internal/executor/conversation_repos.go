package executor

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/yesz1520/autowonder-switch/internal/harness"
)

func appendTool(ctx context.Context, platform harness.PlatformTool, tool json.RawMessage) (string, error) {
	var list struct {
		Tools []json.RawMessage `json:"tools"`
	}
	if platform != nil {
		data, err := platform(ctx, "", nil)
		if err != nil {
			return "", err
		}
		if err = json.Unmarshal([]byte(data), &list); err != nil {
			return "", err
		}
	}
	list.Tools = append(list.Tools, tool)
	b, err := json.Marshal(list)
	return string(b), err
}

func (r *runtime) openConversationRepos(ctx context.Context, pkg, dir string, spec *harness.Request) (func(), error) {
	z, err := zip.OpenReader(pkg)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	f, err := z.Open("repos.json")
	if errors.Is(err, fs.ErrNotExist) {
		return func() {}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var repos []repository
	if err = json.NewDecoder(f).Decode(&repos); err != nil {
		return nil, err
	}
	if len(repos) == 0 {
		return func() {}, nil
	}
	index := make(map[string]repository, len(repos))
	for _, repo := range repos {
		key := strings.ToLower(repo.Name)
		if !validName(repo.Name) || index[key].Name != "" {
			return nil, errors.New("invalid or duplicate conversation repository")
		}
		index[key] = repo
	}
	base, err := os.MkdirTemp(dir, "turn-repositories-")
	if err != nil {
		return nil, err
	}
	closeRepos := func() { os.RemoveAll(base) }
	ready := map[string]string{}
	var mu sync.Mutex
	open := func(ctx context.Context, name string) (string, error) {
		mu.Lock()
		defer mu.Unlock()
		repo := index[strings.ToLower(name)]
		if repo.Name == "" {
			return "", errors.New("repository is not bound to this turn")
		}
		if path := ready[repo.Name]; path != "" {
			return path, nil
		}
		// Stage separately so a failed clone can be retried within this turn.
		temp, err := os.MkdirTemp(base, "checkout-")
		if err != nil {
			return "", err
		}
		checkouts, err := prepareRepoList(ctx, []repository{repo}, temp, filepath.Join(temp, "files"), r.sources, permissions{})
		if err != nil {
			os.RemoveAll(temp)
			return "", err
		}
		ready[repo.Name] = checkouts[0].workDir
		return checkouts[0].workDir, nil
	}
	for _, repo := range repos {
		if repo.Mode == "lazy" {
			continue
		}
		path, err := open(ctx, repo.Name)
		if err != nil {
			closeRepos()
			return nil, err
		}
		spec.Instructions += fmt.Sprintf("\nRepository %q is checked out at %q.", repo.Name, path)
	}
	spec.Readable = true
	spec.ReadOnlyPaths = append(spec.ReadOnlyPaths, base)
	spec.Instructions += "\nUse executor.open_repository with the bound name to locate or check out a repository for this turn."
	platform := spec.Platform
	spec.Platform = func(ctx context.Context, name string, args map[string]any) (string, error) {
		switch name {
		case "":
			return appendTool(ctx, platform, json.RawMessage(`{"name":"executor.open_repository","description":"Open a repository bound to this conversation turn. Returns its local working directory.","inputSchema":{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false}}`))
		case "executor.open_repository":
			repo, _ := args["name"].(string)
			path, err := open(ctx, repo)
			if err != nil {
				return "", err
			}
			b, err := json.Marshal(map[string]string{"path": path})
			return string(b), err
		default:
			if platform == nil {
				return "", errors.New("unknown platform tool")
			}
			return platform(ctx, name, args)
		}
	}
	return closeRepos, nil
}
