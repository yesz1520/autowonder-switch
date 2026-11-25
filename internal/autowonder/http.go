package autowonder

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ErrPlatformTransport denotes an HTTP exchange with no usable response.
var ErrPlatformTransport = errors.New("platform HTTP request failed")

type HTTPError struct {
	StatusCode int
	Body       json.RawMessage
}

func (e *HTTPError) Error() string { return fmt.Sprintf("AutoWonder returned HTTP %d", e.StatusCode) }

type DownloadLink struct {
	DownloadURL      string `json:"downloadUrl"`
	ExpiresInSeconds int64  `json:"expiresInSeconds"`
}
type CheckpointLink struct {
	DownloadLink
	CheckpointSeq int64  `json:"checkpointSeq"`
	ExpiresAt     string `json:"expiresAt"`
	SHA256        string `json:"sha256"`
}
type RecoveryClaim struct {
	Allowed              bool              `json:"allowed"`
	Status               string            `json:"status"`
	EnvironmentVariables map[string]string `json:"environmentVariables"`
}
type CheckpointReceipt struct {
	CheckpointSeq int64  `json:"checkpointSeq"`
	SHA256        string `json:"sha256"`
	SizeBytes     int64  `json:"sizeBytes"`
	Purpose       string `json:"purpose"`
	ResultID      string `json:"resultId"`
}
type FileReceipt struct {
	Index     int    `json:"index"`
	Path      string `json:"path"`
	Status    string `json:"status"`
	Code      string `json:"code"`
	SizeBytes int64  `json:"sizeBytes"`
	MaxBytes  int64  `json:"maxBytes"`
	OSSRef    string `json:"ossRef"`
	RemoteRef string `json:"remoteRef"`
}
type ArtifactsReceipt struct {
	RemoteRef string        `json:"remoteRef"`
	Files     []FileReceipt `json:"files"`
}

func (c *Client) post(ctx context.Context, id int64, action string, params url.Values, contentType string, body io.Reader, out any) error {
	if id <= 0 {
		return errors.New("dispatch ID must be positive")
	}
	if params == nil {
		params = make(url.Values)
	}
	params.Set("token", c.token)
	_, err := c.request(ctx, http.MethodPost, dispatchPath(id, action), params, "", contentType, body, out)
	return err
}

func dispatchPath(id int64, action string) string {
	return "api/daemon/dispatches/" + strconv.FormatInt(id, 10) + "/" + action
}

func (c *Client) request(ctx context.Context, method, path string, params url.Values, bearer, contentType string, body io.Reader, out any) (http.Header, error) {
	u := c.base.JoinPath(path)
	u.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, errors.New("invalid platform request")
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := c.hook.DoHTTP(&c.http, req, path)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, ErrPlatformTransport
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: data}
	}
	if out != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return nil, err
		}
	}
	return resp.Header, nil
}

func (c *Client) jsonRequest(ctx context.Context, method, path string, q url.Values, bearer string, input, out any) error {
	var body io.Reader
	var typ string
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
		typ = "application/json"
	}
	_, err := c.request(ctx, method, path, q, bearer, typ, body, out)
	return err
}

func (c *Client) RefreshPackage(ctx context.Context, id int64) (v DownloadLink, err error) {
	err = c.post(ctx, id, "package-url", nil, "", nil, &v)
	return
}
func (c *Client) RefreshCheckpoint(ctx context.Context, id, seq int64) (v CheckpointLink, err error) {
	if seq <= 0 {
		return v, errors.New("checkpoint sequence must be positive")
	}
	err = c.post(ctx, id, "checkpoint-url", url.Values{"checkpointSeq": {strconv.FormatInt(seq, 10)}}, "", nil, &v)
	if err == nil && v.CheckpointSeq != seq {
		err = errors.New("checkpoint receipt has different sequence")
	}
	return
}
func (c *Client) Claim(ctx context.Context, id int64) (v RecoveryClaim, err error) {
	err = c.post(ctx, id, "recovery-claim", nil, "", nil, &v)
	return
}

// UploadFile is consumed during the request. Its owner closes the reader.
type UploadFile struct {
	Name   string
	Reader io.Reader
}
type CheckpointUpload struct {
	File                                                                    UploadFile
	Seq                                                                     int64
	Purpose, ResultID, Provider, ProviderSessionID, RuntimeID, ActiveStepID string
}

// Interleave multipart headers with the original readers to stream archives
// without buffering their contents or running a separate pipe writer.
func form(field string, files []UploadFile, values map[string]string) (string, io.Reader, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range values {
		if v != "" {
			if err := w.WriteField(k, v); err != nil {
				return "", nil, err
			}
		}
	}
	var parts []io.Reader
	for _, f := range files {
		if f.Name == "" || f.Reader == nil || strings.ContainsAny(f.Name, "\r\n") {
			return "", nil, errors.New("invalid multipart file")
		}
		if _, err := w.CreateFormFile(field, f.Name); err != nil {
			return "", nil, err
		}
		parts = append(parts, bytes.NewReader(bytes.Clone(buf.Bytes())), f.Reader)
		buf.Reset()
	}
	if err := w.Close(); err != nil {
		return "", nil, err
	}
	parts = append(parts, &buf)
	return w.FormDataContentType(), io.MultiReader(parts...), nil
}
func (c *Client) UploadCheckpoint(ctx context.Context, id int64, in CheckpointUpload) (v CheckpointReceipt, err error) {
	if in.Seq < 1 || (in.Purpose != "" && in.Purpose != "RECOVERY" && in.Purpose != "COMPLETION") || (in.Purpose == "COMPLETION") != (in.ResultID != "") {
		return v, errors.New("invalid checkpoint identity")
	}
	typ, body, err := form("checkpoint", []UploadFile{in.File}, map[string]string{"purpose": in.Purpose, "resultId": in.ResultID, "checkpointSeq": strconv.FormatInt(in.Seq, 10), "provider": in.Provider, "providerSessionId": in.ProviderSessionID, "runtimeId": in.RuntimeID, "activeStepId": in.ActiveStepID})
	if err != nil {
		return v, err
	}
	err = c.post(ctx, id, "checkpoint", nil, typ, body, &v)
	if err == nil && (v.CheckpointSeq != in.Seq || v.SHA256 == "" || v.ResultID != in.ResultID || (in.Purpose == "COMPLETION" && v.Purpose != "COMPLETION")) {
		err = errors.New("checkpoint receipt does not match upload")
	}
	return
}
func (c *Client) UploadArtifacts(ctx context.Context, id int64, files []UploadFile, key string, metadata []json.RawMessage) (v ArtifactsReceipt, err error) {
	return c.uploadFiles(ctx, id, "artifacts", files, key, metadata)
}

func (c *Client) UploadDiagnostics(ctx context.Context, id int64, files []UploadFile, key string, metadata []json.RawMessage) (ArtifactsReceipt, error) {
	return c.uploadFiles(ctx, id, "diagnostics", files, key, metadata)
}

func (c *Client) uploadFiles(ctx context.Context, id int64, endpoint string, files []UploadFile, key string, metadata []json.RawMessage) (v ArtifactsReceipt, err error) {
	if len(files) == 0 || len(files) > 200 {
		return v, errors.New("expected 1–200 files")
	}
	fields := map[string]string{"idempotencyKey": key}
	if metadata != nil {
		if len(metadata) != len(files) {
			return v, errors.New("metadata must match file count")
		}
		b, err := json.Marshal(metadata)
		if err != nil {
			return v, err
		}
		fields["filesMetadata"] = string(b)
	}
	typ, body, err := form("files", files, fields)
	if err != nil {
		return v, err
	}
	if err = c.post(ctx, id, endpoint, nil, typ, body, &v); err != nil {
		return
	}
	if len(v.Files) != len(files) {
		return v, errors.New("incomplete artifact receipts")
	}
	seen := make([]bool, len(files))
	for _, f := range v.Files {
		if f.Index < 0 || f.Index >= len(files) || seen[f.Index] {
			return v, errors.New("invalid artifact receipt index")
		}
		seen[f.Index] = true
		if f.Status != "ACCEPTED" {
			return v, fmt.Errorf("artifact %d was not accepted (%s)", f.Index, f.Code)
		}
	}
	return
}

// DownloadVerified installs dest only after the complete signed-URL response
// matches the supplied SHA-256. It never adds the executor credential.
func (c *Client) DownloadVerified(ctx context.Context, link, digest, dest string) error {
	return c.downloadVerified(ctx, link, digest, dest, nil)
}
func (c *Client) downloadVerified(ctx context.Context, link, digest, dest string, check func(string) error) error {
	want, err := hex.DecodeString(strings.TrimPrefix(digest, "sha256:"))
	if err != nil || len(want) != sha256.Size {
		return errors.New("invalid SHA-256")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil || req.URL.User != nil || req.URL.Host == "" || (req.URL.Scheme != "http" && req.URL.Scheme != "https") {
		return errors.New("invalid download URL")
	}
	resp, err := c.hook.DoHTTP(&c.http, req, "signed download")
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return errors.New("file download failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return &HTTPError{StatusCode: resp.StatusCode}
	}
	tmp, err := os.CreateTemp(filepath.Dir(dest), ".download-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	hash := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(tmp, hash), resp.Body)
	err = tmp.Close()
	if copyErr != nil {
		return copyErr
	}
	if err != nil {
		return err
	}
	if !bytes.Equal(hash.Sum(nil), want) {
		return errors.New("SHA-256 mismatch")
	}
	if check != nil {
		if err := check(tmp.Name()); err != nil {
			return err
		}
	}
	return os.Rename(tmp.Name(), dest)
}
