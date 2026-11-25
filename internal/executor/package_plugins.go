package executor

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/yesz1520/autowonder-switch/internal/harness"
)

var pluginName = regexp.MustCompile(`^[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)*$`)

type packagePlugin struct {
	Name, Path string
	Providers  []string
	Required   *bool
}

func (r *runtime) openPackageCapabilities(pkg string, spec *harness.Request) (func(), error) {
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
	var doc struct{ Skills, Plugins []packagePlugin }
	if err = json.NewDecoder(f).Decode(&doc); err != nil {
		return nil, err
	}
	var market, skills string
	cleanup := func() {
		if skills != "" {
			os.RemoveAll(skills)
		}
		if market != "" {
			os.RemoveAll(market)
		}
	}
	if len(doc.Skills) > 0 {
		skills, err = os.MkdirTemp(filepath.Dir(pkg), "turn-skills-")
		if err != nil {
			return nil, err
		}
		if err = unpackSkills(&z.Reader, skills, doc.Skills, spec); err != nil {
			cleanup()
			return nil, err
		}
	}
	seen := map[string]bool{}
	for _, p := range doc.Plugins {
		if !slices.Contains(p.Providers, r.provider()) {
			continue
		}
		if r.provider() != "codex" {
			if p.Required == nil || *p.Required {
				cleanup()
				return nil, fmt.Errorf("required %s package plugin unsupported: %s", r.provider(), p.Name)
			}
			continue
		}
		if !pluginName.MatchString(p.Name) || p.Path != "capabilities/plugins/"+p.Name || seen[strings.ToLower(p.Name)] {
			cleanup()
			return nil, errors.New("invalid or duplicate Codex plugin descriptor")
		}
		seen[strings.ToLower(p.Name)] = true
		if !r.files.NativePlugins {
			if p.Required == nil || *p.Required {
				cleanup()
				return nil, fmt.Errorf("Codex plugin %s requires -native-package-plugins", p.Name)
			}
			continue
		}
		if market == "" {
			home := os.Getenv("CODEX_HOME")
			if home == "" {
				home, err = os.UserHomeDir()
				if err != nil {
					return nil, err
				}
				home = filepath.Join(home, ".codex")
			}
			cache := filepath.Join(home, "plugins", "cache")
			if err = os.MkdirAll(cache, 0700); err != nil {
				return nil, err
			}
			market, err = os.MkdirTemp(cache, "autowonder-")
			if err != nil {
				return nil, err
			}
		}
		if err = unpackPlugin(&z.Reader, market, p); err != nil {
			if p.Required == nil || *p.Required {
				cleanup()
				return nil, fmt.Errorf("plugin %s: %w", p.Name, err)
			}
			spec.Instructions += "\nOptional Codex plugin unavailable: " + p.Name
			continue
		}
		spec.Readable = true
		spec.PluginIDs = append(spec.PluginIDs, p.Name+"@"+filepath.Base(market))
	}
	return cleanup, nil
}

func unpackPlugin(z *zip.Reader, market string, p packagePlugin) error {
	dir := filepath.Join(market, p.Name, "local")
	if err := unpackCapability(z, p.Path, dir); err != nil {
		return err
	}
	var manifest struct{ Name string }
	ok, err := cached(filepath.Join(dir, ".codex-plugin", "plugin.json"), &manifest)
	if err != nil {
		return err
	}
	if !ok || manifest.Name != p.Name {
		return errors.New("missing or mismatched .codex-plugin/plugin.json")
	}
	return nil
}

func unpackCapability(z *zip.Reader, prefix, dir string) error {
	var total uint64
	var entries int
	for _, f := range z.File {
		name, ok := strings.CutPrefix(f.Name, prefix+"/")
		if !ok || f.FileInfo().IsDir() {
			continue
		}
		entries++
		if entries > 500 || !fs.ValidPath(name) || strings.ContainsAny(name, "\\:") || !f.Mode().IsRegular() || f.UncompressedSize64 > 50<<20-total {
			return errors.New("invalid capability entry or bundle exceeds 50 MiB")
		}
		total += f.UncompressedSize64
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return err
		}
		in, err := f.Open()
		if err != nil {
			return err
		}
		mode := os.FileMode(0600)
		if f.Mode().Perm()&0111 != 0 {
			mode = 0700
		}
		out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
		if err != nil {
			in.Close()
			return err
		}
		_, err = io.Copy(out, in)
		if err = errors.Join(err, in.Close(), out.Close()); err != nil {
			return err
		}
	}
	return nil
}

func unpackSkills(z *zip.Reader, dir string, skills []packagePlugin, spec *harness.Request) error {
	seen := map[string]bool{}
	for _, s := range skills {
		if !validName(s.Name) || s.Path != "capabilities/skills/"+s.Name || seen[strings.ToLower(s.Name)] {
			return errors.New("invalid or duplicate packaged skill")
		}
		seen[strings.ToLower(s.Name)] = true
		dest := filepath.Join(dir, s.Name)
		err := unpackCapability(z, s.Path, dest)
		path := filepath.Join(dest, "SKILL.md")
		if err == nil {
			_, err = os.Stat(path)
		}
		if err != nil {
			if s.Required != nil && !*s.Required {
				continue
			}
			return fmt.Errorf("skill %s: %w", s.Name, err)
		}
		spec.Readable = true
		spec.Instructions += fmt.Sprintf("\nPackaged skill %q: read its instructions at %q when relevant. Its scripts and resources are alongside SKILL.md.", s.Name, path)
	}
	return nil
}
