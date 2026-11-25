package autowonder

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"
)

// Event describes one transport exchange. Payloads, credentials, query strings,
// signed URLs and remote error text are deliberately excluded.
type Event struct {
	Time           time.Time `json:"time"`
	Direction      string    `json:"direction"`
	Operation      string    `json:"operation"`
	Outcome        string    `json:"outcome"`
	Code           int       `json:"code,omitempty"`
	DispatchID     int64     `json:"dispatchId,string,omitempty"`
	ConversationID int64     `json:"conversationId,string,omitempty"`
	TurnID         int64     `json:"turnId,string,omitempty"`
	RequestID      string    `json:"requestId,omitempty"`
	ResultID       string    `json:"resultId,omitempty"`
	Phase          string    `json:"phase,omitempty"`
	Status         string    `json:"status,omitempty"`
}

// EventHook can be called concurrently by sessions sharing a client.
type EventHook func(Event)

func (h EventHook) emit(e Event) {
	if h != nil {
		e.Time = time.Now().UTC()
		h(e)
	}
}

// DoHTTP records the HTTP response status. Operation must be a trusted route
// or a static description, never a credential-bearing URL.
func (h EventHook) DoHTTP(client *http.Client, req *http.Request, operation string) (*http.Response, error) {
	resp, err := client.Do(req)
	if h != nil {
		e := Event{Direction: "http", Operation: req.Method + " " + operation, Outcome: "failed"}
		if resp != nil {
			e.Code = resp.StatusCode
			if err == nil && e.Code >= 200 && e.Code < 300 {
				e.Outcome = "ok"
			}
		}
		h.emit(e)
	}
	return resp, err
}

var eventName = regexp.MustCompile("^[A-Z][A-Z0-9_]{0,79}$")
var eventID = regexp.MustCompile("^[a-zA-Z0-9_.:-]{1,100}$")

func (h EventHook) frame(direction string, data []byte, err error) {
	if h == nil {
		return
	}
	var meta struct {
		Type, Status, ResultID, Phase      string
		RequestID                          json.RawMessage
		DispatchID, ConversationID, TurnID int64
	}
	_ = json.Unmarshal(data, &meta)
	e := Event{Direction: direction, Operation: "UNKNOWN_FRAME", Outcome: "ok", DispatchID: meta.DispatchID, ConversationID: meta.ConversationID, TurnID: meta.TurnID}
	if eventName.MatchString(meta.Type) {
		e.Operation = meta.Type
	}
	if eventName.MatchString(meta.Status) {
		e.Status = meta.Status
	}
	if eventID.MatchString(meta.ResultID) {
		e.ResultID = meta.ResultID
	}
	if eventID.MatchString(meta.Phase) {
		e.Phase = meta.Phase
	}
	var id string
	if json.Unmarshal(meta.RequestID, &id) == nil && eventID.MatchString(id) {
		e.RequestID = id
	}
	if err != nil {
		e.Outcome = "failed"
	}
	h.emit(e)
}
