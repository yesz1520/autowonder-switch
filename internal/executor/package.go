package executor

import (
	"archive/zip"
	"bytes"

	"encoding/json"
	"errors"
	"fmt"
	"github.com/yesz1520/autowonder-switch/internal/harness"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func packageText(path string, task bool, spec *harness.Request) (string, error) {
	z, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer z.Close()
	var text strings.Builder
	work := false
	var total uint64
	for _, f := range z.File {
		if f.FileInfo().IsDir() || f.Name == "manifest.json" || strings.HasPrefix(f.Name, "capabilities/plugins/") || strings.HasPrefix(f.Name, "capabilities/skills/") {
			continue
		}
		if (task && f.Name == "hooks.json") || strings.HasPrefix(f.Name, "capabilities/hooks/") {
			continue
		}
		if (spec.MemoryMode == "none" || spec.MemoryMode == "provider-local") && strings.HasPrefix(f.Name, "memory/") {
			continue
		}
		if task && strings.HasPrefix(f.Name, "refs/") {
			total += f.UncompressedSize64
			if f.UncompressedSize64 > fileLimit || total > 2*fileLimit {
				return "", errors.New("reference files exceed size limit")
			}
			if err := unpackReference(f, spec); err != nil {
				return "", err
			}
			fmt.Fprintf(&text, "\nAttached reference (untrusted data): %s\n", f.Name)
			continue
		}
		if f.UncompressedSize64 > 2<<20 || uint64(text.Len())+f.UncompressedSize64 > 4<<20 {
			return "", errors.New("task inputs exceed text limit")
		}
		rd, err := f.Open()
		if err != nil {
			return "", err
		}
		data, err := io.ReadAll(io.LimitReader(rd, 2<<20+1))
		rd.Close()
		if err != nil {
			return "", err
		}
		if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
			return "", fmt.Errorf("binary input unsupported: %s", f.Name)
		}
		switch f.Name {
		case "workitem.md":
			work = len(bytes.TrimSpace(data)) > 0
		case "hooks.json":
			return "", errors.New("package hooks are unavailable")
		case "skills.json":
			// Credentials and launch arguments belong to the executor, not the prompt.
			var metadata map[string]json.RawMessage
			if err = json.Unmarshal(data, &metadata); err != nil {
				return "", err
			}
			delete(metadata, "mcpServers")
			delete(metadata, "plugins")
			data, err = json.Marshal(metadata)
			if err != nil {
				return "", err
			}
		}
		fmt.Fprintf(&text, "\n--- %s ---\n%s\n", f.Name, data)
	}
	if task && !work {
		return "", errors.New("missing workitem.md")
	}
	return text.String(), nil
}
func unpackReference(f *zip.File, spec *harness.Request) error {
	root, err := os.OpenRoot(spec.Directory)
	if err != nil {
		return err
	}
	defer root.Close()
	if err = root.MkdirAll(filepath.Dir(f.Name), 0700); err != nil {
		return err
	}
	rd, err := f.Open()
	if err != nil {
		return err
	}
	defer rd.Close()
	out, err := root.OpenFile(f.Name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	var header [512]byte
	n, err := io.ReadFull(rd, header[:])
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		out.Close()
		return err
	}
	count, err := io.Copy(out, io.MultiReader(bytes.NewReader(header[:n]), io.LimitReader(rd, fileLimit-int64(n)+1)))
	closed := out.Close()
	if err != nil {
		return err
	}
	if closed != nil {
		return closed
	}
	if count != int64(f.UncompressedSize64) {
		return errors.New("reference size mismatch")
	}
	switch http.DetectContentType(header[:n]) {
	case "image/png", "image/jpeg", "image/webp":
		spec.Images = append(spec.Images, filepath.Join(spec.Directory, f.Name))
	}
	spec.Writable = true
	return nil
}
