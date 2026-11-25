package autowonder

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
)

// Manifest uses the ID strings sent inside the archive (WebSocket IDs are integers).
type Manifest struct {
	SchemaVersion        string            `json:"schemaVersion"`
	PackageID            string            `json:"packageId"`
	DispatchID           string            `json:"dispatchId"`
	SourceDispatchID     string            `json:"sourceDispatchId"`
	TenantID             string            `json:"tenantId"`
	WorkitemID           string            `json:"workitemId"`
	ExecutorID           string            `json:"executorId"`
	AgentID              string            `json:"agentId"`
	AgentVersionID       string            `json:"agentVersionId"`
	SDLCID               string            `json:"sdlcId"`
	SDLCStepID           string            `json:"sdlcStepId"`
	Attempt              int               `json:"attempt"`
	IdempotencyKey       string            `json:"idempotencyKey"`
	CapabilityTreeDigest string            `json:"capabilityTreeDigest"`
	FileDigests          map[string]string `json:"fileDigests"`
}

// ReadTaskPackage checks the manifest and every file digest without executing
// or extracting anything in the archive.
func ReadTaskPackage(path string) (Manifest, error) {
	var m Manifest
	z, err := zip.OpenReader(path)
	if err != nil {
		return m, err
	}
	defer z.Close()
	files := map[string]*zip.File{}
	for _, f := range z.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if !fs.ValidPath(f.Name) || strings.Contains(f.Name, "\\") || files[f.Name] != nil || !f.Mode().IsRegular() {
			return m, errors.New("invalid task package entry")
		}
		files[f.Name] = f
	}
	f := files["manifest.json"]
	if f == nil {
		return m, errors.New("missing task package manifest")
	}
	r, err := f.Open()
	if err != nil {
		return m, err
	}
	defer r.Close()
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		return m, err
	}
	delete(files, "manifest.json")
	if m.SchemaVersion != "autoWonder.taskPackage.v1" || m.FileDigests == nil || len(files) != len(m.FileDigests) {
		return m, errors.New("invalid task package manifest")
	}
	for name, f := range files {
		want, ok := m.FileDigests[name]
		if !ok {
			return m, errors.New("unlisted task package entry")
		}
		r, err := f.Open()
		if err != nil {
			return m, err
		}
		h := sha256.New()
		_, err = io.Copy(h, r)
		r.Close()
		if err != nil {
			return m, err
		}
		if fmt.Sprintf("sha256:%x", h.Sum(nil)) != want {
			return m, fmt.Errorf("digest mismatch for %q", name)
		}
	}
	return m, nil
}
