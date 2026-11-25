# Acknowledge turn delivery

English | [中文](conversation-delivery-ack.zh.md)

Platform → executor. Frame type: `CONVERSATION_DELIVERY_ACK`.

Match the original frame by `conversationDeliveryProtocolVersion: 1`, `frameType`, `conversationId`, `turnId` and `dispatchAttempt`. For events also match `eventSeq` and `chunkIndex`.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `CONVERSATION_DELIVERY_ACK`. |
| `conversationDeliveryProtocolVersion` | Integer | Turn receipts enabled | Receipt protocol version for turn events and final replies. |
| `frameType` | String |  | Type of the frame being acknowledged. |
| `conversationId` | Integer |  | Conversation that receives or produces the message. |
| `turnId` | Integer |  | Turn ID within the conversation. |
| `dispatchAttempt` | Integer | For the specified turn attempt | Turn delivery attempt number. Echo the supplied value. |
| `eventSeq` | Integer | Every receipt | Event sequence number within the turn. Normally zero when acknowledging the final reply. |
| `chunkIndex` | Integer | Every receipt | Zero-based chunk index. Normally zero when acknowledging the final reply. |

## Example frame

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
- [CONVERSATION_DELIVERY_ACK_V1](../features/conversation-delivery-ack-v1.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
