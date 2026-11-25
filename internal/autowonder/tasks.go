package autowonder

import "encoding/json"

type TaskDispatch struct {
	ExecutorID                 int64                 `json:"executorId,omitzero"`
	DispatchID                 int64                 `json:"dispatchId,omitzero"`
	TenantID                   int64                 `json:"tenantId,omitzero"`
	WorkitemID                 int64                 `json:"workitemId,omitzero"`
	AgentID                    int64                 `json:"agentId,omitzero"`
	AgentVersionID             int64                 `json:"agentVersionId,omitzero"`
	Attempt                    int64                 `json:"attempt,omitzero"`
	DownloadURL                string                `json:"downloadUrl,omitzero"`
	Size                       int64                 `json:"size,omitzero"`
	ChecksumAlgorithm          string                `json:"checksumAlgorithm,omitzero"`
	ChecksumScope              string                `json:"checksumScope,omitzero"`
	Checksum                   string                `json:"checksum,omitzero"`
	ArtifactUploadPath         string                `json:"artifactUploadPath,omitzero"`
	CheckpointUploadPath       string                `json:"checkpointUploadPath,omitzero"`
	PackageRefreshPath         string                `json:"packageRefreshPath,omitzero"`
	SDLCStepID                 int64                 `json:"sdlcStepId,omitzero"`
	IdempotencyKey             string                `json:"idempotencyKey,omitzero"`
	MD5                        string                `json:"md5,omitzero"`
	PackageID                  string                `json:"packageId,omitzero"`
	Issuer                     string                `json:"issuer,omitzero"`
	SignatureRef               string                `json:"signatureRef,omitzero"`
	Signature                  string                `json:"signature,omitzero"`
	SignatureAlgorithm         string                `json:"signatureAlgorithm,omitzero"`
	SignaturePublicKey         string                `json:"signaturePublicKey,omitzero"`
	ExpiresAt                  string                `json:"expiresAt,omitzero"`
	AllowCommit                *bool                 `json:"allowCommit,omitzero"`
	AllowPush                  *bool                 `json:"allowPush,omitzero"`
	AllowNetwork               *bool                 `json:"allowNetwork,omitzero"`
	CheckpointRefreshPath      string                `json:"checkpointRefreshPath,omitzero"`
	ResumeMode                 string                `json:"resumeMode,omitzero"`
	ResumeSessionBehavior      string                `json:"resumeSessionBehavior,omitzero"`
	ResumeFromDispatchID       int64                 `json:"resumeFromDispatchId,omitzero"`
	ResumeProvider             string                `json:"resumeProvider,omitzero"`
	ResumeSessionID            string                `json:"resumeSessionId,omitzero"`
	ResumeCheckpointURL        string                `json:"resumeCheckpointUrl,omitzero"`
	ResumeCheckpointSHA256     string                `json:"resumeCheckpointSha256,omitzero"`
	ResumeCheckpointSeq        int64                 `json:"resumeCheckpointSeq,omitzero"`
	ResumeCheckpointCandidates []CheckpointCandidate `json:"resumeCheckpointCandidates,omitzero"`
	MemorySnapshotPath         string                `json:"memorySnapshotPath,omitzero"`
	MemoryChangesPath          string                `json:"memoryChangesPath,omitzero"`
	MemoryMutationPath         string                `json:"memoryMutationPath,omitzero"`
	MemoryRecallPath           string                `json:"memoryRecallPath,omitzero"`
	MemoryStartingRevisions    map[string]int64      `json:"memoryStartingRevisions,omitzero"`
	ModelSelection             *ModelSelection       `json:"modelSelection,omitzero"`
	DispatchMCPToken           string                `json:"dispatchMcpToken,omitzero"`
	MCPSecrets                 map[string]string     `json:"mcpSecrets,omitzero"`
	EnvironmentVariables       map[string]string     `json:"environmentVariables,omitzero"`
	DebugLog                   json.RawMessage       `json:"debugLog,omitzero"`
}

func (*TaskDispatch) Kind() string { return "TASK_DISPATCH" }
func (*TaskDispatch) incoming()    {}

type TaskAck struct {
	DispatchID int64 `json:"dispatchId,omitzero"`
}

func (*TaskAck) Kind() string { return "TASK_ACK" }
func (*TaskAck) outgoing()    {}

type TaskBusy struct {
	DispatchID int64  `json:"dispatchId,omitzero"`
	Reason     string `json:"reason,omitzero"`
}

func (*TaskBusy) Kind() string { return "TASK_BUSY" }
func (*TaskBusy) outgoing()    {}

type TaskProgress struct {
	ResultSummary string          `json:"resultSummary,omitempty"`
	RuntimeEvent  json.RawMessage `json:"runtimeEvent,omitzero"`
	Detail        json.RawMessage `json:"detail,omitzero"`
	Log           string          `json:"log,omitempty"`
	EventID       string          `json:"eventId,omitempty"`
	Seq           *int64          `json:"seq,omitempty"`
	DispatchID    int64           `json:"dispatchId,omitzero"`
	EventType     string          `json:"eventType,omitzero"`
	Message       string          `json:"message,omitzero"`
	StepOrder     int64           `json:"stepOrder,omitzero"`
	StepName      string          `json:"stepName,omitzero"`
}

func (*TaskProgress) Kind() string { return "TASK_PROGRESS" }
func (*TaskProgress) outgoing()    {}

type TaskResult struct {
	Retryable             *bool           `json:"retryable,omitempty"`
	DispatchID            int64           `json:"dispatchId,omitzero"`
	ResultProtocolVersion int64           `json:"resultProtocolVersion,omitzero"`
	ResultID              string          `json:"resultId,omitzero"`
	Success               bool            `json:"success"`
	Error                 string          `json:"error,omitzero"`
	ResultSummary         string          `json:"resultSummary,omitzero"`
	CheckpointSeq         int64           `json:"checkpointSeq,omitzero"`
	CheckpointSHA256      string          `json:"checkpointSha256,omitzero"`
	FailureCategory       string          `json:"failureCategory,omitzero"`
	WorkflowChanged       *bool           `json:"workflowChanged,omitzero"`
	Handoff               json.RawMessage `json:"handoff,omitzero"`
	WorkflowPlan          json.RawMessage `json:"workflowPlan,omitzero"`
	DebugLog              json.RawMessage `json:"debugLog,omitzero"`
}

func (*TaskResult) Kind() string { return "TASK_RESULT" }
func (*TaskResult) outgoing()    {}

type TaskResultAck struct {
	DispatchID               int64  `json:"dispatchId,omitzero"`
	ResultProtocolVersion    int64  `json:"resultProtocolVersion,omitzero"`
	ResultID                 string `json:"resultId,omitzero"`
	Accepted                 bool   `json:"accepted"`
	Disposition              string `json:"disposition,omitzero"`
	Retryable                *bool  `json:"retryable,omitzero"`
	ReasonCode               string `json:"reasonCode,omitzero"`
	PauseCheckpointConfirmed *bool  `json:"pauseCheckpointConfirmed,omitzero"`
	CheckpointSeq            int64  `json:"checkpointSeq,omitzero"`
	CheckpointSHA256         string `json:"checkpointSha256,omitzero"`
}

func (*TaskResultAck) Kind() string { return "TASK_RESULT_ACK" }
func (*TaskResultAck) incoming()    {}

type TaskPause struct {
	ExecutorID int64 `json:"executorId,omitzero"`
	DispatchID int64 `json:"dispatchId,omitzero"`
	Cancel     *bool `json:"cancel,omitzero"`
}

func (*TaskPause) Kind() string { return "TASK_PAUSE" }
func (*TaskPause) incoming()    {}

type TaskPaused struct {
	DispatchID       int64  `json:"dispatchId,omitzero"`
	CheckpointSeq    int64  `json:"checkpointSeq,omitzero"`
	CheckpointSHA256 string `json:"checkpointSha256,omitzero"`
}

func (*TaskPaused) Kind() string { return "TASK_PAUSED" }
func (*TaskPaused) outgoing()    {}

type TaskPauseFailed struct {
	DispatchID int64  `json:"dispatchId,omitzero"`
	Error      string `json:"error,omitzero"`
}

func (*TaskPauseFailed) Kind() string { return "TASK_PAUSE_FAILED" }
func (*TaskPauseFailed) outgoing()    {}

type TaskHandoff struct {
	DispatchID int64  `json:"dispatchId,omitzero"`
	WorkitemID int64  `json:"workitemId,omitzero"`
	To         string `json:"to,omitzero"`
	ToType     string `json:"toType,omitzero"`
}

func (*TaskHandoff) Kind() string { return "TASK_HANDOFF" }
func (*TaskHandoff) outgoing()    {}

type TaskHandoffResult struct {
	DispatchID           int64  `json:"dispatchId,omitzero"`
	WorkitemID           int64  `json:"workitemId,omitzero"`
	Status               string `json:"status,omitzero"`
	DownstreamDispatchID int64  `json:"downstreamDispatchId,omitzero"`
	TargetType           string `json:"targetType,omitzero"`
	TargetRef            int64  `json:"targetRef,omitzero"`
	Message              string `json:"message,omitzero"`
	ReasonCode           string `json:"reasonCode,omitzero"`
}

func (*TaskHandoffResult) Kind() string { return "TASK_HANDOFF_RESULT" }
func (*TaskHandoffResult) incoming()    {}

type TaskGuidance struct {
	ExecutorID    int64  `json:"executorId,omitzero"`
	GuidanceID    int64  `json:"guidanceId,omitzero"`
	DispatchID    int64  `json:"dispatchId,omitzero"`
	WorkitemID    int64  `json:"workitemId,omitzero"`
	TargetAgentID int64  `json:"targetAgentId,omitzero"`
	Content       string `json:"content,omitzero"`
	Mode          string `json:"mode,omitzero"`
}

func (*TaskGuidance) Kind() string { return "TASK_GUIDANCE" }
func (*TaskGuidance) incoming()    {}

type TaskGuidanceAck struct {
	GuidanceID    int64           `json:"guidanceId,omitzero"`
	DispatchID    int64           `json:"dispatchId,omitzero"`
	Status        string          `json:"status,omitzero"`
	ReplyMarkdown string          `json:"replyMarkdown,omitzero"`
	Error         string          `json:"error,omitzero"`
	WorkflowPlan  json.RawMessage `json:"workflowPlan,omitzero"`
}

func (*TaskGuidanceAck) Kind() string { return "TASK_GUIDANCE_ACK" }
func (*TaskGuidanceAck) outgoing()    {}

type TaskSteer struct {
	ExecutorID int64  `json:"executorId,omitzero"`
	GuidanceID int64  `json:"guidanceId,omitzero"`
	DispatchID int64  `json:"dispatchId,omitzero"`
	Content    string `json:"content,omitzero"`
}

func (*TaskSteer) Kind() string { return "TASK_STEER" }
func (*TaskSteer) incoming()    {}

type TaskSteerAck struct {
	DispatchID int64  `json:"dispatchId,omitzero"`
	GuidanceID int64  `json:"guidanceId,omitzero"`
	Status     string `json:"status,omitzero"`
	Error      string `json:"error,omitzero"`
}

func (*TaskSteerAck) Kind() string { return "TASK_STEER_ACK" }
func (*TaskSteerAck) outgoing()    {}

type ArtifactUploaded struct {
	DispatchID   int64  `json:"dispatchId,omitzero"`
	Name         string `json:"name,omitzero"`
	ArtifactType string `json:"artifactType,omitzero"`
	OSSRef       string `json:"ossRef,omitzero"`
	Size         int64  `json:"size,omitzero"`
	MetaJSON     string `json:"metaJson,omitzero"`
}

func (*ArtifactUploaded) Kind() string { return "ARTIFACT_UPLOADED" }
func (*ArtifactUploaded) outgoing()    {}

type WorkitemWorkspaceCleanup struct {
	ExecutorID      int64 `json:"executorId,omitzero"`
	TenantID        int64 `json:"tenantId,omitzero"`
	WorkitemID      int64 `json:"workitemId,omitzero"`
	WorkitemVersion int64 `json:"workitemVersion,omitzero"`
	PublishedAt     int64 `json:"publishedAt,omitzero"`
}

func (*WorkitemWorkspaceCleanup) Kind() string { return "WORKITEM_WORKSPACE_CLEANUP" }
func (*WorkitemWorkspaceCleanup) incoming()    {}
