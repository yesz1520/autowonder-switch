// Package autowonder implements the executor side of the AutoWonder 1.3.0 wire protocol.
package autowonder

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

// Incoming and Outgoing distinguish the two directions at compile time.
// The transport supplies the type discriminator; callers supply the payload.
type Incoming interface {
	Kind() string
	incoming()
}
type Outgoing interface {
	Kind() string
	outgoing()
}

type Model struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type ModelSelection struct {
	Model         string `json:"model"`
	DefaultModel  string `json:"defaultModel"`
	EntryCount    int    `json:"entryCount"`
	Threshold     *int   `json:"threshold"`
	PolicyVersion int    `json:"policyVersion"`
	ReasonCode    string `json:"reasonCode"`
}
type CheckpointCandidate struct {
	DownloadURL   string `json:"downloadUrl"`
	SHA256        string `json:"sha256"`
	CheckpointSeq int64  `json:"checkpointSeq"`
}

type LegacyConversationTurn struct {
	TurnID int64 `json:"turnId"`
}

type Heartbeat struct {
	RunningConversationTurns     []LegacyConversationTurn `json:"runningConversationTurns,omitzero"`
	StartedAt                    string                   `json:"startedAt,omitempty"`
	Version                      string                   `json:"version,omitempty"`
	Model                        string                   `json:"model,omitempty"`
	RestartRequestID             string                   `json:"restartRequestId,omitempty"`
	ProtocolFeatures             []Feature                `json:"protocolFeatures,omitempty"`
	MaxConcurrentDispatches      int                      `json:"maxConcurrentDispatches,omitempty"`
	DispatchInventoryReady       *bool                    `json:"dispatchInventoryReady,omitempty"`
	DispatchInventoryError       string                   `json:"dispatchInventoryError,omitempty"`
	RunningDispatchIDs           []int64                  `json:"runningDispatchIds,omitzero"`
	OwnedDispatchIDs             []int64                  `json:"ownedDispatchIds,omitzero"`
	RunningConversationTurnIDs   []int64                  `json:"runningConversationTurnIds,omitzero"`
	ComputingConversationTurnIDs []int64                  `json:"computingConversationTurnIds,omitzero"`
}

func (*Heartbeat) Kind() string { return "HEARTBEAT" }
func (*Heartbeat) outgoing()    {}

func Encode(m Outgoing) ([]byte, error) {
	if m == nil {
		return nil, errors.New("missing outgoing message")
	}
	if h, ok := m.(*Heartbeat); ok && h != nil {
		if err := h.validate(); err != nil {
			return nil, err
		}
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var payload map[string]json.RawMessage
	if err = json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if payload == nil {
		return nil, errors.New("nil message payload")
	}
	kind, err := json.Marshal(m.Kind())
	if err != nil {
		return nil, err
	}
	payload["type"] = kind
	return json.Marshal(payload)
}
func (h *Heartbeat) validate() error {
	if len(h.ProtocolFeatures) > 64 {
		return errors.New("heartbeat exceeds 64 features")
	}
	if slices.Contains(h.ProtocolFeatures, FeatureDispatchInventoryV1) {
		if h.MaxConcurrentDispatches < 1 || h.MaxConcurrentDispatches > 50 || h.DispatchInventoryReady == nil {
			return errors.New("inventory requires capacity 1–50 and readiness")
		}
		if err := validIDs("runningDispatchIds", h.RunningDispatchIDs, 50); err != nil {
			return err
		}
		if err := validIDs("ownedDispatchIds", h.OwnedDispatchIDs, 1000); err != nil {
			return err
		}
		if err := validIDs("runningConversationTurnIds", h.RunningConversationTurnIDs, 50); err != nil {
			return err
		}
		for _, id := range h.RunningDispatchIDs {
			if !slices.Contains(h.OwnedDispatchIDs, id) {
				return fmt.Errorf("running dispatch %d is not owned", id)
			}
		}
		switch h.DispatchInventoryError {
		case "":
		case "OWNED_DISPATCH_LIMIT_EXCEEDED":
			if *h.DispatchInventoryReady || len(h.OwnedDispatchIDs) != len(h.RunningDispatchIDs) {
				return errors.New("overflow requires incomplete inventory and only running IDs")
			}
		default:
			return errors.New("unknown inventory error")
		}
	}
	if slices.Contains(h.ProtocolFeatures, FeatureConversationComputeInventoryV1) {
		if err := validIDs("computingConversationTurnIds", h.ComputingConversationTurnIDs, 50); err != nil {
			return err
		}
		for _, id := range h.ComputingConversationTurnIDs {
			if !slices.Contains(h.RunningConversationTurnIDs, id) {
				return fmt.Errorf("computing turn %d is not owned", id)
			}
		}
	}
	return nil
}
func validIDs(field string, ids []int64, max int) error {
	if ids == nil || len(ids) > max {
		return fmt.Errorf("%s requires an array of at most %d IDs", field, max)
	}
	seen := make(map[int64]bool, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			return fmt.Errorf("invalid or duplicate ID in %s", field)
		}
		seen[id] = true
	}
	return nil
}
