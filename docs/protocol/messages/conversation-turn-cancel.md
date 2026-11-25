# Cancel a turn

English | [中文](conversation-turn-cancel.zh.md)

Platform → executor. Frame type: `CONVERSATION_TURN_CANCEL`.

Cancel the turn identified by `executorId`, `conversationId` and `turnId`. If `dispatchAttempt` is present, cancel only that attempt.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `CONVERSATION_TURN_CANCEL`. |
| `executorId` | Integer |  | ID of the executor on this connection. |
| `conversationId` | Integer |  | Conversation that receives or produces the message. |
| `turnId` | Integer |  | Turn ID within the conversation. |
| `dispatchAttempt` | Integer | For the specified turn attempt | Turn delivery attempt number. Echo the supplied value. |

## Example frame

```json
{
  "type": "CONVERSATION_TURN_CANCEL",
  "executorId": 42,
  "conversationId": 80,
  "turnId": 81,
  "dispatchAttempt": 1
}
```

## Example reply

See [CONVERSATION_TURN_ACK](conversation-turn-ack.md) for the reply fields.

```json
{
  "type": "CONVERSATION_TURN_ACK",
  "conversationId": 80,
  "turnId": 81,
  "dispatchAttempt": 1,
  "status": "CANCELED",
  "sessionId": "session-1"
}
```

## Related protocol

- [CONVERSATION_TURN_ACK](conversation-turn-ack.md)
- [CONVERSATION_TURN_CANCEL](../features/conversation-turn-cancel.md)

Message handling in [WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java).
