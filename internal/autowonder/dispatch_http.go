package autowonder

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"net/http"
	"net/url"
	"strconv"
)

func (c *Client) dispatchJSON(ctx context.Context, id int64, method, route string, q url.Values, input, out any) error {
	if id < 1 {
		return errors.New("dispatch ID must be positive")
	}
	if q == nil {
		q = make(url.Values)
	}
	q.Set("token", c.token)
	return c.jsonRequest(ctx, method, dispatchPath(id, route), q, "", input, out)
}

func (c *Client) Environment(ctx context.Context) (map[string]string, error) {
	var response struct {
		EnvironmentVariables map[string]string `json:"environmentVariables"`
	}
	err := c.jsonRequest(ctx, http.MethodGet, "api/daemon/executors/"+strconv.FormatInt(c.id, 10)+"/environment-variables", nil, c.token, nil, &response)
	return response.EnvironmentVariables, err
}

type ExternalLink struct {
	Provider           string `json:"provider"`
	ExternalProjectID  string `json:"externalProjectId"`
	ExternalWorkitemID string `json:"externalWorkitemId"`
	ExternalURL        string `json:"externalUrl"`
}
type WorkitemIdentity struct {
	WorkitemID    int64          `json:"workitemId"`
	ExternalLinks []ExternalLink `json:"externalLinks"`
}

func (c *Client) WorkitemIdentity(ctx context.Context, id int64) (v WorkitemIdentity, err error) {
	err = c.dispatchJSON(ctx, id, http.MethodGet, "workitem-identity", nil, nil, &v)
	return
}

type TaskDataWrite struct {
	Namespace       string          `json:"namespace"`
	Key             string          `json:"key"`
	Value           json.RawMessage `json:"value"`
	OwnerWorkitemID *int64          `json:"ownerWorkitemId,omitempty"`
	ExpectedVersion *int            `json:"expectedVersion,omitempty"`
}
type TaskDataRecord struct {
	ID              int64           `json:"id"`
	Namespace       string          `json:"namespace"`
	RecordKey       string          `json:"recordKey"`
	Value           json.RawMessage `json:"value"`
	OwnerWorkitemID int64           `json:"ownerWorkitemId"`
	Version         int             `json:"version"`
	CreatedAt       json.RawMessage `json:"gmtCreate"`
	ModifiedAt      json.RawMessage `json:"gmtModified"`
}
type TaskDataPage struct {
	Items       []TaskDataRecord `json:"items"`
	NextAfterID *int64           `json:"nextAfterId"`
}

func (c *Client) GetTaskData(ctx context.Context, id int64, namespace, key string) (v TaskDataRecord, err error) {
	if key == "" {
		return v, errors.New("record key is required")
	}
	err = c.dispatchJSON(ctx, id, http.MethodGet, "task-data", url.Values{"namespace": {namespace}, "key": {key}}, nil, &v)
	return
}
func (c *Client) ListTaskData(ctx context.Context, id int64, namespace string, after int64, limit int) (v TaskDataPage, err error) {
	q := url.Values{"namespace": {namespace}}
	if after != 0 {
		q.Set("afterId", strconv.FormatInt(after, 10))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	err = c.dispatchJSON(ctx, id, http.MethodGet, "task-data", q, nil, &v)
	return
}
func (c *Client) PutTaskData(ctx context.Context, id int64, input TaskDataWrite) (v TaskDataRecord, err error) {
	err = c.dispatchJSON(ctx, id, http.MethodPut, "task-data", nil, input, &v)
	return
}

type Comment struct {
	ContentMD      string  `json:"contentMd"`
	TargetHumanIDs []int64 `json:"targetHumanIds,omitempty"`
}

// Comment and status responses use the platform's work-item / scheduled-run view.
func (c *Client) AddComment(ctx context.Context, id int64, input Comment) (v json.RawMessage, err error) {
	err = c.dispatchJSON(ctx, id, http.MethodPost, "comments", nil, input, &v)
	return
}
func (c *Client) SetWorkitemStatus(ctx context.Context, id int64, status string) (v json.RawMessage, err error) {
	err = c.dispatchJSON(ctx, id, http.MethodPost, "workitem-status", nil, map[string]string{"status": status}, &v)
	return
}

type AcceptanceEvidence struct {
	Kind  string `json:"kind"`
	Ref   string `json:"ref"`
	Title string `json:"title,omitempty"`
}
type AcceptanceReport struct {
	RequirementKey      string               `json:"requirementKey"`
	Implementation      string               `json:"implementation"`
	Verdict             string               `json:"verdict"`
	Note                string               `json:"note,omitempty"`
	NotApplicableReason string               `json:"notApplicableReason,omitempty"`
	Evidence            []AcceptanceEvidence `json:"evidence"`
}
type AcceptanceReceipt struct {
	DispatchID int64 `json:"dispatchId"`
	Reported   int   `json:"reported"`
}

func (c *Client) ReportAcceptance(ctx context.Context, id int64, input []AcceptanceReport) (v AcceptanceReceipt, err error) {
	err = c.dispatchJSON(ctx, id, http.MethodPost, "acceptance", nil, input, &v)
	return
}

type HumanRequestInput struct {
	Kind           string          `json:"kind"`
	Key            string          `json:"key"`
	StepID         string          `json:"stepId"`
	StepAttempt    int             `json:"stepAttempt"`
	Message        string          `json:"message"`
	Schema         json.RawMessage `json:"schema,omitempty"`
	Context        json.RawMessage `json:"context,omitempty"`
	TimeoutSeconds int             `json:"timeoutSeconds"`
}
type HumanRequest struct {
	ID               int64           `json:"id"`
	DispatchID       int64           `json:"dispatchId"`
	Kind             string          `json:"kind"`
	Key              string          `json:"key"`
	Body             json.RawMessage `json:"body"`
	Status           string          `json:"status"`
	Answer           json.RawMessage `json:"answer"`
	ActorID          *int64          `json:"actorId"`
	ExpiresAt        json.RawMessage `json:"expiresAt"`
	SuspendRequested bool            `json:"suspendRequested"`
	ResumeDispatchID *int64          `json:"resumeDispatchId"`
	AnsweredAt       json.RawMessage `json:"answeredAt"`
}

func (c *Client) CreateHumanRequest(ctx context.Context, id int64, input HumanRequestInput) (v HumanRequest, err error) {
	err = c.jsonRequest(ctx, http.MethodPost, dispatchPath(id, "requests"), nil, c.token, input, &v)
	return
}
func (c *Client) GetHumanRequest(ctx context.Context, id, requestID int64) (v HumanRequest, err error) {
	err = c.jsonRequest(ctx, http.MethodGet, dispatchPath(id, "requests/"+strconv.FormatInt(requestID, 10)), nil, c.token, nil, &v)
	return
}
func (c *Client) CancelHumanRequest(ctx context.Context, id, requestID int64) error {
	var response struct {
		Status string `json:"status"`
	}
	if err := c.jsonRequest(ctx, http.MethodPost, dispatchPath(id, "requests/"+strconv.FormatInt(requestID, 10)+"/cancel"), nil, c.token, nil, &response); err != nil {
		return err
	}
	if response.Status != "CANCELED" {
		return errors.New("human request cancellation not confirmed")
	}
	return nil
}

type UsageEntry struct {
	Provider         string   `json:"provider"`
	Model            string   `json:"model"`
	StepID           string   `json:"step_id,omitempty"`
	InputTokens      *int64   `json:"input_tokens,omitempty"`
	OutputTokens     *int64   `json:"output_tokens,omitempty"`
	CacheReadTokens  *int64   `json:"cache_read_tokens,omitempty"`
	CacheWriteTokens *int64   `json:"cache_write_tokens,omitempty"`
	ReasoningTokens  *int64   `json:"reasoning_tokens,omitempty"`
	Credits          *float64 `json:"credits,omitempty"`
}

func (c *Client) ReportUsage(ctx context.Context, id int64, entries []UsageEntry) error {
	if id < 1 {
		return errors.New("dispatch ID must be positive")
	}
	return c.jsonRequest(ctx, http.MethodPost, "api/daemon/tasks/"+strconv.FormatInt(id, 10)+"/usage", nil, c.token, struct {
		Usage []UsageEntry `json:"usage"`
	}{entries}, nil)
}

type DebugLogUploadRequest struct {
	SizeBytes      int64  `json:"sizeBytes"`
	SHA256         string `json:"sha256"`
	Truncated      bool   `json:"truncated"`
	DispatchStatus string `json:"dispatchStatus"`
}
type DebugLogUploadLink struct {
	ObjectKey       string `json:"objectKey"`
	UploadURL       string `json:"uploadUrl"`
	ExpiresAt       string `json:"expiresAt"`
	AlreadyUploaded bool   `json:"alreadyUploaded"`
}

func (c *Client) DebugLogUpload(ctx context.Context, id int64, input DebugLogUploadRequest) (v DebugLogUploadLink, err error) {
	err = c.dispatchJSON(ctx, id, http.MethodPost, "debug-log-upload", nil, input, &v)
	return
}

type LogArchiveCapability struct {
	Supported       bool
	MaxDecodedBytes int64
}

func (c *Client) LogArchiveCapability(ctx context.Context, id int64) (v LogArchiveCapability, err error) {
	h, err := c.request(ctx, http.MethodOptions, dispatchPath(id, "artifacts"), url.Values{"token": {c.token}}, "", "", nil, nil)
	if err != nil {
		return v, err
	}
	v.Supported = h.Get("X-AutoWonder-Log-Archive") == "1"
	if v.Supported {
		v.MaxDecodedBytes, err = strconv.ParseInt(h.Get("X-AutoWonder-Log-Archive-Max-Decoded-Bytes"), 10, 64)
	}
	return
}

// PutDebugLog sends the bytes to the signed object-storage destination. The
// caller supplies the length used in DebugLogUpload; executor auth is never added.
func (c *Client) PutDebugLog(ctx context.Context, link DebugLogUploadLink, body io.Reader, size int64) error {
	if link.AlreadyUploaded {
		return nil
	}
	if body == nil || size < 0 {
		return errors.New("missing log data or invalid length")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, link.UploadURL, body)
	if err != nil || req.URL.User != nil || req.URL.Host == "" || (req.URL.Scheme != "https" && req.URL.Scheme != "http") {
		return errors.New("invalid debug log upload URL")
	}
	req.ContentLength = size
	resp, err := c.hook.DoHTTP(&c.http, req, "signed debug log upload")
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return errors.New("debug log upload failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{StatusCode: resp.StatusCode}
	}
	return nil
}
