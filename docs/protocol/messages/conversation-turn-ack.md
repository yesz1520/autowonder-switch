# Submit a final turn reply

English | [中文](conversation-turn-ack.zh.md)

Executor → platform. Frame type: `CONVERSATION_TURN_ACK`.

Return the final status for `conversationId`, `turnId` and the supplied `dispatchAttempt`. Send `SUCCESS` with `replyMarkdown` and `sessionId`, `FAILED` with `error`, or `CANCELED` after cancellation.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `CONVERSATION_TURN_ACK`. |
| `conversationId` | Integer |  | Conversation that receives or produces the message. |
| `turnId` | Integer |  | Turn ID within the conversation. |
| `dispatchAttempt` | Integer | For the specified turn attempt | Turn delivery attempt number. Echo the supplied value. |
| `status` | String | When the turn ends | Turn outcome: SUCCESS, FAILED or CANCELED. |
| `replyMarkdown` | String | When returning a reply | Reply body in Markdown. |
| `sessionId` | String |  | Native session ID used for the turn. |
| `conversationDeliveryProtocolVersion` | Integer | Turn receipts enabled | Receipt protocol version for turn events and final replies. |
| `usage` | Object | When reporting model usage | Model usage for the turn. |
| `error` | String | On failure | Description of the failure. |

## Example frame

```json
{
  "type": "CONVERSATION_TURN_ACK",
  "conversationId": 80,
  "turnId": 81,
  "dispatchAttempt": 1,
  "status": "SUCCESS",
  "replyMarkdown": "Review complete",
  "sessionId": "session-1",
  "conversationDeliveryProtocolVersion": 1
}
```

## Example reply

See [CONVERSATION_DELIVERY_ACK](conversation-delivery-ack.md) for the reply fields.

```json
{
  "type": "CONVERSATION_DELIVERY_ACK",
  "conversationDeliveryProtocolVersion": 1,
  "frameType": "CONVERSATION_TURN_ACK",
  "conversationId": 80,
  "turnId": 81,
  "dispatchAttempt": 1,
  "eventSeq": 0,
  "chunkIndex": 0
}
```

## Related protocol

- [Conversation turns](../conversations.md)
- [CONVERSATION_DELIVERY_ACK](conversation-delivery-ack.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
