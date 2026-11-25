package autowonder

import (
	"encoding/json"
	"fmt"
)

// Unknown messages retain the original JSON so newer servers can add frame types.
type Unknown struct {
	Name string
	Data json.RawMessage
}

func (u *Unknown) Kind() string { return u.Name }
func (*Unknown) incoming()      {}

func Decode(data []byte) (Incoming, error) {
	return decodeForExecutor(data, 0)
}

func decodeForExecutor(data []byte, expectedID int64) (Incoming, error) {
	var header struct {
		ExecutorID *int64 `json:"executorId"`
		Type       string `json:"type"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return nil, err
	}
	if expectedID > 0 && header.ExecutorID != nil && *header.ExecutorID != expectedID {
		return nil, fmt.Errorf("message addressed to another executor")
	}
	var msg Incoming
	switch header.Type {
	case "":
		return nil, fmt.Errorf("missing message type")
	case "TASK_DISPATCH":
		msg = new(TaskDispatch)
	case "TASK_RESULT_ACK":
		msg = new(TaskResultAck)
	case "TASK_PAUSE":
		msg = new(TaskPause)
	case "TASK_HANDOFF_RESULT":
		msg = new(TaskHandoffResult)
	case "TASK_GUIDANCE":
		msg = new(TaskGuidance)
	case "TASK_STEER":
		msg = new(TaskSteer)
	case "CONVERSATION_TURN":
		msg = new(ConversationTurn)
	case "CONVERSATION_TURN_CANCEL":
		msg = new(ConversationTurnCancel)
	case "CONVERSATION_DELIVERY_ACK":
		msg = new(ConversationDeliveryAck)
	case "CONVERSATION_COMMANDS_PROBE":
		msg = new(ConversationCommandsProbe)
	case "CONVERSATION_ELICITATION_REPLY":
		msg = new(ConversationElicitationReply)
	case "QODER_MODEL_CATALOG_REQUEST":
		msg = new(QoderModelCatalogRequest)
	case "MCP_CONNECTION_TEST":
		msg = new(MCPConnectionTest)
	case "WORKITEM_WORKSPACE_CLEANUP":
		msg = new(WorkitemWorkspaceCleanup)
	case "EXECUTOR_RESTART":
		msg = new(ExecutorRestart)
	case "EXECUTOR_UPGRADE":
		msg = new(ExecutorUpgrade)
	case "EXECUTOR_UPGRADE_RESULT_SUPPORT":
		msg = new(ExecutorUpgradeResultSupport)
	case "EXECUTOR_UPGRADE_RESULT_ACK":
		msg = new(ExecutorUpgradeResultAck)
	default:
		return &Unknown{header.Type, append(json.RawMessage(nil), data...)}, nil
	}
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, fmt.Errorf("%s: %w", header.Type, err)
	}
	return msg, nil
}
