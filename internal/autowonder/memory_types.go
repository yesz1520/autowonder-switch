package autowonder

import "encoding/json"

type MemoryImportManifestRequest struct {
	ProviderFamily          string `json:"providerFamily"`
	LogicalPath             string `json:"logicalPath"`
	InstallationFingerprint string `json:"installationFingerprint"`
	SanitizedContentSHA256  string `json:"sanitizedContentSha256"`
	ByteSize                int64  `json:"byteSize"`
	ObservationStatus       string `json:"observationStatus"`
}

type MemoryImportSnapshotRequest struct {
	SourceID               int64  `json:"sourceId"`
	SanitizedContentSHA256 string `json:"sanitizedContentSha256"`
	SanitizedContent       string `json:"sanitizedContent"`
	ByteSize               int64  `json:"byteSize"`
	RedactionCount         int    `json:"redactionCount"`
}

type MemoryImportCompletionRequest struct {
	SourceID               int64  `json:"sourceId"`
	SanitizedContentSHA256 string `json:"sanitizedContentSha256"`
	RedactionCount         int    `json:"redactionCount"`
	LeaseID                string `json:"leaseId"`
}

type MemoryImportClaimRequest struct {
	SourceID               int64  `json:"sourceId"`
	SanitizedContentSHA256 string `json:"sanitizedContentSha256"`
	LeaseID                string `json:"leaseId"`
}

type MemoryMaintenanceLeaseRequest struct {
	StoreID int64  `json:"storeId"`
	LeaseID string `json:"leaseId"`
}

type MemoryMutationRequest struct {
	Operation          string `json:"operation"`
	Path               string `json:"path"`
	NewPath            string `json:"newPath"`
	ContentMD          string `json:"contentMd"`
	ExpectedVersion    *int   `json:"expectedVersion"`
	IdempotencyKey     string `json:"idempotencyKey"`
	MaintenanceLeaseID string `json:"maintenanceLeaseId"`
}

type MemoryChangePage struct {
	StoreID         int64          `json:"storeId"`
	CurrentRevision int64          `json:"currentRevision"`
	Changes         []MemoryChange `json:"changes"`
	HasMore         bool           `json:"hasMore"`
}

type MemoryImportDecision struct {
	UploadRequired         bool   `json:"uploadRequired"`
	SourceID               int64  `json:"sourceId"`
	Reason                 string `json:"reason"`
	CurationRequired       bool   `json:"curationRequired"`
	SanitizedContentSHA256 string `json:"sanitizedContentSha256"`
}

type MemoryImportClaim struct {
	Acquired               bool   `json:"acquired"`
	SourceID               int64  `json:"sourceId"`
	SanitizedContentSHA256 string `json:"sanitizedContentSha256"`
	LeaseID                string `json:"leaseId"`
	CurrentContent         string `json:"currentContent"`
	PreviousContent        string `json:"previousContent"`
	RedactionCount         int    `json:"redactionCount"`
}

type MemoryMaintenanceLease struct {
	Acquired bool   `json:"acquired"`
	LeaseID  string `json:"leaseId"`
}

type MemoryDocument struct {
	ID               *int64          `json:"id"`
	TenantID         *int64          `json:"tenantId"`
	StoreID          *int64          `json:"storeId"`
	Path             string          `json:"path"`
	ContentMD        string          `json:"contentMd"`
	ContentSHA256    string          `json:"contentSha256"`
	ByteSize         *int64          `json:"byteSize"`
	ModifiedAt       json.RawMessage `json:"modifiedAt"`
	DeletedAt        json.RawMessage `json:"deletedAt"`
	CreatorType      string          `json:"creatorType"`
	CreatorRef       string          `json:"creatorRef"`
	SourceDispatchID *int64          `json:"sourceDispatchId"`
	SourceSessionID  string          `json:"sourceSessionId"`
	CreatorID        *int64          `json:"creatorId"`
	ModifierID       *int64          `json:"modifierId"`
	Version          *int            `json:"version"`
	GmtCreate        json.RawMessage `json:"gmtCreate"`
	GmtModified      json.RawMessage `json:"gmtModified"`
	Title            string          `json:"title"`
	MemoryType       string          `json:"memoryType"`
	Description      string          `json:"description"`
	DescriptionKind  string          `json:"descriptionKind"`
	Body             string          `json:"body"`
}

type MemoryChange struct {
	ID                 *int64          `json:"id"`
	TenantID           *int64          `json:"tenantId"`
	StoreID            *int64          `json:"storeId"`
	DocumentID         *int64          `json:"documentId"`
	StoreRevision      *int64          `json:"storeRevision"`
	Operation          string          `json:"operation"`
	Path               string          `json:"path"`
	DocumentVersion    *int            `json:"documentVersion"`
	ContentSHA256      string          `json:"contentSha256"`
	ActorType          string          `json:"actorType"`
	ActorRef           string          `json:"actorRef"`
	DispatchID         *int64          `json:"dispatchId"`
	ProviderSessionID  string          `json:"providerSessionId"`
	IdempotencyKey     string          `json:"idempotencyKey"`
	RequestFingerprint string          `json:"requestFingerprint"`
	GmtCreate          json.RawMessage `json:"gmtCreate"`
}
