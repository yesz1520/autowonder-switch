# Report a turn event

English | [中文](conversation-turn-event.zh.md)

Executor → platform. Frame type: `CONVERSATION_TURN_EVENT`.

Identify the conversation, turn and attempt. Send `eventType`, `eventSeq`, zero-based `chunkIndex`, `chunkCount` and `payloadFragment` for each event fragment.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `CONVERSATION_TURN_EVENT`. |
| `conversationId` | Integer |  | Conversation that receives or produces the message. |
| `turnId` | Integer |  | Turn ID within the conversation. |
| `dispatchAttempt` | Integer | For the specified turn attempt | Turn delivery attempt number. Echo the supplied value. |
| `eventType` | String |  | Event kind used to interpret its payload. |
| `eventSeq` | Integer | Chunked event | Event sequence number within the turn. |
| `chunkIndex` | Integer | Chunked event | Zero-based chunk index. |
| `chunkCount` | Integer | Chunked event | Total number of chunks for the event. |
| `payloadFragment` | String | Chunked event | Serialized event content carried by this chunk. |
| `conversationDeliveryProtocolVersion` | Integer | Turn receipts enabled | Receipt protocol version for turn events and final replies. |

## Example frame

```json
{
  "type": "CONVERSATION_TURN_EVENT",
  "conversationId": 80,
  "turnId": 81,
  "dispatchAttempt": 1,
  "eventType": "status",
  "eventSeq": 1,
  "chunkIndex": 0,
  "chunkCount": 1,
  "payloadFragment": "{\"status\":\"running\"}",
  "conversationDeliveryProtocolVersion": 1
}
```

## Example reply

See [CONVERSATION_DELIVERY_ACK](conversation-delivery-ack.md) for the reply fields.

```json
{
  "type": "CONVERSATION_DELIVERY_ACK",
  "conversationDeliveryProtocolVersion": 1,
  "frameType": "CONVERSATION_TURN_EVENT",
  "conversationId": 80,
  "turnId": 81,
  "dispatchAttempt": 1,
  "eventSeq": 1,
  "chunkIndex": 0
}
```

## Related protocol

- [Conversation turns](../conversations.md)
- [CONVERSATION_DELIVERY_ACK](conversation-delivery-ack.md)
- [CONVERSATION_TURN_EVENT](../features/conversation-turn-event.md)
- [CONVERSATION_ACP_INTERACTION_V1](../features/conversation-acp-interaction-v1.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
