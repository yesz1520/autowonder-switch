# Turn events

English | [中文](conversation-turn-event.zh.md)

Heartbeat capability identifier: `CONVERSATION_TURN_EVENT`.

Report events before the turn finishes. Each event has a sequence number and its payload can span multiple chunks.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "CONVERSATION_TURN_EVENT"
  ]
}
```

## Related protocol

- [CONVERSATION_TURN_EVENT](../messages/conversation-turn-event.md)

Capability implementation in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
