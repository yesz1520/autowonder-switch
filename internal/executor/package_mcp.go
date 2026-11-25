package executor

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"strings"

	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/mcpclient"
)

var mcpName = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// openPackageTools keeps credentials in the executor. Only the tool catalog
// and tool results reach the model; sessions live for this dispatch or turn.
func (r *runtime) openPackageTools(ctx context.Context, pkg string, secrets map[string]string, spec *harness.Request) (func(), error) {
	z, err := zip.OpenReader(pkg)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	f, err := z.Open("skills.json")
	if errors.Is(err, fs.ErrNotExist) {
		return func() {}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var doc struct{ MCPServers []mcpclient.Server }
	if err = json.NewDecoder(f).Decode(&doc); err != nil {
		return nil, err
	}
	clients := map[string]*mcpclient.Session{}
	var catalog []json.RawMessage
	tools := map[string]*mcpclient.Session{}
	closeAll := func() {
		for _, s := range clients {
			s.Close()
		}
	}
	for _, server := range doc.MCPServers {
		if server.Name == "autowonder" {
			if server.Required && spec.Platform == nil {
				closeAll()
				return nil, errors.New("missing platform MCP token")
			}
			continue
		}
		if !mcpName.MatchString(server.Name) || clients[server.Name] != nil {
			closeAll()
			return nil, errors.New("invalid or duplicate package MCP name")
		}
		session, err := mcpclient.Open(ctx, server, secrets, spec.Directory, r.codexBinary)
		if err != nil {
			if server.Required {
				closeAll()
				return nil, fmt.Errorf("MCP %s: %w", server.Name, err)
			}
			spec.Instructions += "\nOptional MCP service unavailable: " + server.Name
			continue
		}
		clients[server.Name] = session
		for _, t := range session.Tools {
			qualified := server.Name + "/" + t.Name
			info := *t
			info.Name = qualified
			b, err := json.Marshal(info)
			if err != nil {
				closeAll()
				return nil, err
			}
			catalog = append(catalog, b)
			tools[qualified] = session
		}
	}
	if len(clients) == 0 {
		return closeAll, nil
	}
	platform := spec.Platform
	spec.Platform = func(ctx context.Context, name string, args map[string]any) (string, error) {
		if name == "" {
			var list struct {
				Tools []json.RawMessage `json:"tools"`
			}
			if platform != nil {
				b, err := platform(ctx, "", nil)
				if err != nil {
					return "", err
				}
				if err = json.Unmarshal([]byte(b), &list); err != nil {
					return "", err
				}
			}
			list.Tools = append(list.Tools, catalog...)
			b, err := json.Marshal(list)
			return string(b), err
		}
		if session := tools[name]; session != nil {
			_, tool, _ := strings.Cut(name, "/")
			v, err := session.Call(ctx, tool, args)
			if err != nil {
				return "", errors.New("package MCP call failed")
			}
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			if v.IsError {
				return "", fmt.Errorf("package MCP tool returned an error: %s", b)
			}
			return string(b), nil
		}
		if strings.Contains(name, "/") || platform == nil {
			return "", errors.New("unknown MCP tool")
		}
		return platform(ctx, name, args)
	}
	return closeAll, nil
}
