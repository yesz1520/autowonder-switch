package autowonder

// Matches distinguishes V2 result receipts from pause and legacy receipts.
// A match identifies a report; consult Disposition before deleting or retrying it.
func (a *TaskResultAck) Matches(result *TaskResult) bool {
	return result != nil && a.ResultProtocolVersion == 2 && result.ResultProtocolVersion == 2 && a.DispatchID == result.DispatchID && a.ResultID != "" && a.ResultID == result.ResultID
}
func (a *ConversationDeliveryAck) MatchesEvent(e *ConversationTurnEvent) bool {
	return e != nil && a.ConversationDeliveryProtocolVersion == 1 && e.ConversationDeliveryProtocolVersion == 1 && a.FrameType == e.Kind() && a.ConversationID == e.ConversationID && a.TurnID == e.TurnID && a.DispatchAttempt == e.DispatchAttempt && a.EventSeq == e.EventSeq && a.ChunkIndex == e.ChunkIndex
}
func (a *ConversationDeliveryAck) MatchesReply(r *ConversationTurnAck) bool {
	return r != nil && a.ConversationDeliveryProtocolVersion == 1 && r.ConversationDeliveryProtocolVersion == 1 && a.FrameType == r.Kind() && a.ConversationID == r.ConversationID && a.TurnID == r.TurnID && a.DispatchAttempt == r.DispatchAttempt
}
