package autowonder

import "encoding/json"

type QoderModelCatalogRequest struct {
	ExecutorID int64  `json:"executorId,omitzero"`
	RequestID  string `json:"requestId,omitzero"`
	Provider   string `json:"provider,omitzero"`
}

func (*QoderModelCatalogRequest) Kind() string { return "QODER_MODEL_CATALOG_REQUEST" }
func (*QoderModelCatalogRequest) incoming()    {}

type QoderModelCatalogResult struct {
	RequestID string  `json:"requestId,omitzero"`
	Provider  string  `json:"provider,omitzero"`
	Success   bool    `json:"success"`
	Models    []Model `json:"models,omitzero"`
}

func (*QoderModelCatalogResult) Kind() string { return "QODER_MODEL_CATALOG_RESULT" }
func (*QoderModelCatalogResult) outgoing()    {}

type MCPConnectionTest struct {
	ExecutorID     int64             `json:"executorId,omitzero"`
	TestID         string            `json:"testId,omitzero"`
	Transport      string            `json:"transport,omitzero"`
	Command        string            `json:"command,omitzero"`
	Args           []string          `json:"args,omitzero"`
	Env            map[string]string `json:"env,omitzero"`
	TimeoutSeconds int64             `json:"timeoutSeconds,omitzero"`
	URL            string            `json:"url,omitzero"`
	Headers        map[string]string `json:"headers,omitzero"`
}

func (*MCPConnectionTest) Kind() string { return "MCP_CONNECTION_TEST" }
func (*MCPConnectionTest) incoming()    {}

type MCPConnectionTestResult struct {
	TestID     string            `json:"testId,omitzero"`
	Success    bool              `json:"success"`
	DurationMS int64             `json:"durationMs,omitzero"`
	Tools      []json.RawMessage `json:"tools,omitzero"`
	Message    string            `json:"message,omitzero"`
}

func (*MCPConnectionTestResult) Kind() string { return "MCP_CONNECTION_TEST_RESULT" }
func (*MCPConnectionTestResult) outgoing()    {}

type ExecutorRestart struct {
	ExecutorID           int64           `json:"executorId,omitzero"`
	RequestID            string          `json:"requestId,omitzero"`
	IssuedAt             string          `json:"issuedAt,omitzero"`
	Update               *bool           `json:"update,omitzero"`
	Status               string          `json:"status,omitzero"`
	RestartMode          string          `json:"restartMode,omitzero"`
	Launch               json.RawMessage `json:"launch,omitzero"`
	LaunchConfigVersion  int64           `json:"launchConfigVersion,omitzero"`
	TargetRuntimeVersion string          `json:"targetRuntimeVersion,omitzero"`
	PreviousStartedAt    int64           `json:"previousStartedAt,omitzero"`
	RequestedBy          int64           `json:"requestedBy,omitzero"`
	MessageOrigin        string          `json:"messageOrigin,omitzero"`
	Message              string          `json:"message,omitzero"`
}

func (*ExecutorRestart) Kind() string { return "EXECUTOR_RESTART" }
func (*ExecutorRestart) incoming()    {}

type ExecutorRestartResult struct {
	RequestID string `json:"requestId,omitzero"`
	Status    string `json:"status,omitzero"`
	Message   string `json:"message,omitzero"`
}

func (*ExecutorRestartResult) Kind() string { return "EXECUTOR_RESTART_RESULT" }
func (*ExecutorRestartResult) outgoing()    {}

type ExecutorUpgrade struct {
	ExecutorID       int64  `json:"executorId,omitzero"`
	RequestID        string `json:"requestId,omitzero"`
	TargetVersion    string `json:"targetVersion,omitzero"`
	IssuedAt         string `json:"issuedAt,omitzero"`
	ResultReceiptAck *bool  `json:"resultReceiptAck,omitzero"`
}

func (*ExecutorUpgrade) Kind() string { return "EXECUTOR_UPGRADE" }
func (*ExecutorUpgrade) incoming()    {}

type ExecutorUpgradeResult struct {
	RequestID      string `json:"requestId,omitzero"`
	Phase          string `json:"phase,omitzero"`
	ReceiptID      string `json:"receiptId,omitzero"`
	CurrentVersion string `json:"currentVersion,omitzero"`
	Error          string `json:"error,omitzero"`
	Terminal       *bool  `json:"terminal,omitzero"`
}

func (*ExecutorUpgradeResult) Kind() string { return "EXECUTOR_UPGRADE_RESULT" }
func (*ExecutorUpgradeResult) outgoing()    {}

type ExecutorUpgradeResultSupport struct {
	RequestID string `json:"requestId,omitzero"`
	ReceiptID string `json:"receiptId,omitzero"`
}

func (*ExecutorUpgradeResultSupport) Kind() string { return "EXECUTOR_UPGRADE_RESULT_SUPPORT" }
func (*ExecutorUpgradeResultSupport) incoming()    {}

type ExecutorUpgradeResultAck struct {
	RequestID string `json:"requestId,omitzero"`
	ReceiptID string `json:"receiptId,omitzero"`
}

func (*ExecutorUpgradeResultAck) Kind() string { return "EXECUTOR_UPGRADE_RESULT_ACK" }
func (*ExecutorUpgradeResultAck) incoming()    {}
