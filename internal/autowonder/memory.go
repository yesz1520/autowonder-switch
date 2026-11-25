package autowonder

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

// Memory is bound to the dispatch-scoped credential, not the executor token.
type Memory struct {
	client *Client
	id     int64
	token  string
}

func (c *Client) Memory(id int64, dispatchMCPToken string) (*Memory, error) {
	if id <= 0 || dispatchMCPToken == "" {
		return nil, errors.New("memory requires a dispatch ID and dispatch MCP token")
	}
	return &Memory{c, id, dispatchMCPToken}, nil
}
func (m *Memory) call(ctx context.Context, method, route string, q url.Values, in, out any) error {
	return m.client.jsonRequest(ctx, method, dispatchPath(m.id, "memory/"+route), q, m.token, in, out)
}

type MemorySnapshot struct {
	ComposedIndex string `json:"composedIndex"`
	Stores        []struct {
		ID         int64  `json:"id"`
		Scope      string `json:"scope"`
		OwnerRef   int64  `json:"ownerRef"`
		Revision   int64  `json:"revision"`
		Permission string `json:"permission"`
		Documents  []struct {
			Path      string `json:"path"`
			ContentMD string `json:"contentMd"`
			Version   int    `json:"version"`
		} `json:"documents"`
	} `json:"stores"`
}

func (m *Memory) Snapshot(ctx context.Context) (v MemorySnapshot, err error) {
	err = m.call(ctx, http.MethodGet, "snapshot", nil, nil, &v)
	return
}
func (m *Memory) Changes(ctx context.Context, storeID, after int64, limit int) (v MemoryChangePage, err error) {
	q := url.Values{"storeId": {strconv.FormatInt(storeID, 10)}, "afterRevision": {strconv.FormatInt(after, 10)}}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	err = m.call(ctx, http.MethodGet, "changes", q, nil, &v)
	return
}
func (m *Memory) Mutate(ctx context.Context, storeID int64, input MemoryMutationRequest) (v MemoryDocument, err error) {
	err = m.call(ctx, http.MethodPost, "mutations", url.Values{"storeId": {strconv.FormatInt(storeID, 10)}}, input, &v)
	return
}
func (m *Memory) Recall(ctx context.Context, storeID int64, path string) error {
	return m.call(ctx, http.MethodPost, "recalls", nil, struct {
		StoreID int64  `json:"storeId"`
		Path    string `json:"path"`
	}{storeID, path}, nil)
}
func (m *Memory) ImportManifest(ctx context.Context, in MemoryImportManifestRequest) (v MemoryImportDecision, err error) {
	err = m.call(ctx, http.MethodPost, "imports/manifest", nil, in, &v)
	return
}
func (m *Memory) ImportSnapshot(ctx context.Context, in MemoryImportSnapshotRequest) error {
	return m.call(ctx, http.MethodPost, "imports/snapshots", nil, in, nil)
}
func (m *Memory) CompleteImport(ctx context.Context, in MemoryImportCompletionRequest) error {
	return m.call(ctx, http.MethodPost, "imports/completions", nil, in, nil)
}
func (m *Memory) ClaimImport(ctx context.Context, in MemoryImportClaimRequest) (v MemoryImportClaim, err error) {
	err = m.call(ctx, http.MethodPost, "imports/claims", nil, in, &v)
	return
}
func (m *Memory) RenewImport(ctx context.Context, in MemoryImportClaimRequest) error {
	return m.call(ctx, http.MethodPost, "imports/renewals", nil, in, nil)
}
func (m *Memory) ReleaseImport(ctx context.Context, in MemoryImportClaimRequest) error {
	return m.call(ctx, http.MethodPost, "imports/releases", nil, in, nil)
}
func (m *Memory) ClaimMaintenance(ctx context.Context, in MemoryMaintenanceLeaseRequest) (v MemoryMaintenanceLease, err error) {
	err = m.call(ctx, http.MethodPost, "maintenance/claims", nil, in, &v)
	return
}
func (m *Memory) RenewMaintenance(ctx context.Context, in MemoryMaintenanceLeaseRequest) error {
	return m.call(ctx, http.MethodPost, "maintenance/renewals", nil, in, nil)
}
func (m *Memory) ReleaseMaintenance(ctx context.Context, in MemoryMaintenanceLeaseRequest) error {
	return m.call(ctx, http.MethodPost, "maintenance/releases", nil, in, nil)
}
