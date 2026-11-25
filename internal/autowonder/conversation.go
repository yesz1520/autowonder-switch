package autowonder

import "encoding/json"

type ConversationTurn struct {
	ExecutorID                          int64             `json:"executorId,omitzero"`
	ConversationID                      int64             `json:"conversationId,omitzero"`
	TurnID                              int64             `json:"turnId,omitzero"`
	AgentID                             int64             `json:"agentId,omitzero"`
	DispatchAttempt                     int64             `json:"dispatchAttempt,omitzero"`
	Content                             string            `json:"content,omitzero"`
	RequestID                           string            `json:"requestId,omitzero"`
	SystemPrompt                        string            `json:"systemPrompt,omitzero"`
	CLISessionRef                       string            `json:"cliSessionRef,omitzero"`
	AgentVersionID                      int64             `json:"agentVersionId,omitzero"`
	CapabilityDownloadURL               string            `json:"capabilityDownloadUrl,omitzero"`
	CapabilitySHA256                    string            `json:"capabilitySha256,omitzero"`
	CapabilityHash                      string            `json:"capabilityHash,omitzero"`
	MCPToken                            string            `json:"mcpToken,omitzero"`
	ConversationDeliveryProtocolVersion int64             `json:"conversationDeliveryProtocolVersion,omitzero"`
	ConversationAnswerProtocolVersion   int64             `json:"conversationAnswerProtocolVersion,omitzero"`
	MCPSecrets                          map[string]string `json:"mcpSecrets,omitzero"`
	EnvironmentVariables                map[string]string `json:"environmentVariables,omitzero"`
	ConversationActorProtocolVersion    int64             `json:"conversationActorProtocolVersion,omitzero"`
}

func (*ConversationTurn) Kind() string { return "CONVERSATION_TURN" }
func (*ConversationTurn) incoming()    {}

type ConversationTurnEvent struct {
	ConversationID                      int64  `json:"conversationId,omitzero"`
	TurnID                              int64  `json:"turnId,omitzero"`
	DispatchAttempt                     int64  `json:"dispatchAttempt,omitzero"`
	EventType                           string `json:"eventType,omitzero"`
	EventSeq                            int64  `json:"eventSeq"`
	ChunkIndex                          int64  `json:"chunkIndex"`
	ChunkCount                          int64  `json:"chunkCount,omitzero"`
	PayloadFragment                     string `json:"payloadFragment,omitzero"`
	ConversationDeliveryProtocolVersion int64  `json:"conversationDeliveryProtocolVersion,omitzero"`
}

func (*ConversationTurnEvent) Kind() string { return "CONVERSATION_TURN_EVENT" }
func (*ConversationTurnEvent) outgoing()    {}

type ConversationTurnCancel struct {
	ExecutorID      int64 `json:"executorId,omitzero"`
	ConversationID  int64 `json:"conversationId,omitzero"`
	TurnID          int64 `json:"turnId,omitzero"`
	DispatchAttempt int64 `json:"dispatchAttempt,omitzero"`
}

func (*ConversationTurnCancel) Kind() string { return "CONVERSATION_TURN_CANCEL" }
func (*ConversationTurnCancel) incoming()    {}

type ConversationTurnAck struct {
	ConversationID                      int64           `json:"conversationId,omitzero"`
	TurnID                              int64           `json:"turnId,omitzero"`
	DispatchAttempt                     int64           `json:"dispatchAttempt,omitzero"`
	Status                              string          `json:"status,omitzero"`
	ReplyMarkdown                       string          `json:"replyMarkdown,omitzero"`
	SessionID                           string          `json:"sessionId,omitzero"`
	ConversationDeliveryProtocolVersion int64           `json:"conversationDeliveryProtocolVersion,omitzero"`
	Usage                               json.RawMessage `json:"usage,omitzero"`
	Error                               string          `json:"error,omitzero"`
}

func (*ConversationTurnAck) Kind() string { return "CONVERSATION_TURN_ACK" }
func (*ConversationTurnAck) outgoing()    {}

type ConversationDeliveryAck struct {
	ConversationDeliveryProtocolVersion int64  `json:"conversationDeliveryProtocolVersion,omitzero"`
	FrameType                           string `json:"frameType,omitzero"`
	ConversationID                      int64  `json:"conversationId,omitzero"`
	TurnID                              int64  `json:"turnId,omitzero"`
	DispatchAttempt                     int64  `json:"dispatchAttempt,omitzero"`
	EventSeq                            int64  `json:"eventSeq"`
	ChunkIndex                          int64  `json:"chunkIndex"`
}

func (*ConversationDeliveryAck) Kind() string { return "CONVERSATION_DELIVERY_ACK" }
func (*ConversationDeliveryAck) incoming()    {}

type ConversationCommandsProbe struct {
	ExecutorID     int64 `json:"executorId,omitzero"`
	ConversationID int64 `json:"conversationId,omitzero"`
}

func (*ConversationCommandsProbe) Kind() string { return "CONVERSATION_COMMANDS_PROBE" }
func (*ConversationCommandsProbe) incoming()    {}

type ConversationCommandsResult struct {
	ConversationID int64  `json:"conversationId,omitzero"`
	Status         string `json:"status,omitzero"`
	Commands       string `json:"commands"`
	Error          string `json:"error,omitzero"`
}

func (*ConversationCommandsResult) Kind() string { return "CONVERSATION_COMMANDS_RESULT" }
func (*ConversationCommandsResult) outgoing()    {}

type ConversationElicitationReply struct {
	ExecutorID            int64           `json:"executorId,omitzero"`
	ConversationID        int64           `json:"conversationId,omitzero"`
	TurnID                int64           `json:"turnId,omitzero"`
	RequestID             string          `json:"requestId,omitzero"`
	Action                string          `json:"action,omitzero"`
	Content               json.RawMessage `json:"content,omitzero"`
	AnswerProtocolVersion int64           `json:"answerProtocolVersion,omitzero"`
	AnswerID              string          `json:"answerId,omitzero"`
	DispatchAttempt       int64           `json:"dispatchAttempt,omitzero"`
	RemainingTTLMS        int64           `json:"remainingTtlMs,omitzero"`
	ReceiptOnly           *bool           `json:"receiptOnly,omitzero"`
}

func (*ConversationElicitationReply) Kind() string { return "CONVERSATION_ELICITATION_REPLY" }
func (*ConversationElicitationReply) incoming()    {}

type ConversationElicitationReplyAck struct {
	ConversationID        int64  `json:"conversationId,omitzero"`
	TurnID                int64  `json:"turnId,omitzero"`
	RequestID             string `json:"requestId,omitzero"`
	AnswerID              string `json:"answerId,omitzero"`
	DispatchAttempt       int64  `json:"dispatchAttempt,omitzero"`
	AnswerProtocolVersion int64  `json:"answerProtocolVersion,omitzero"`
	Status                string `json:"status,omitzero"`
}

func (*ConversationElicitationReplyAck) Kind() string { return "CONVERSATION_ELICITATION_REPLY_ACK" }
func (*ConversationElicitationReplyAck) outgoing()    {}
