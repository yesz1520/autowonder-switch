# Turn cancellation

English | [中文](conversation-turn-cancel.zh.md)

Heartbeat capability identifier: `CONVERSATION_TURN_CANCEL`.

Supports canceling an individual turn. When the command includes `dispatchAttempt`, stop only that attempt.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "CONVERSATION_TURN_CANCEL"
  ]
}
```

## Related protocol

- [CONVERSATION_TURN_CANCEL](../messages/conversation-turn-cancel.md)

Capability implementation in [WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java).
