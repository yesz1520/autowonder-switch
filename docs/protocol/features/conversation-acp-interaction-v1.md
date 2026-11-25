# ACP interactions during a turn

English | [中文](conversation-acp-interaction-v1.zh.md)

Heartbeat capability identifier: `CONVERSATION_ACP_INTERACTION_V1`.

Supports ACP interactions within a turn. The executor raises a question via a turn event and the platform delivers the user’s answer for that turn.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "CONVERSATION_ACP_INTERACTION_V1"
  ]
}
```

## Related protocol

- [CONVERSATION_TURN_EVENT](../messages/conversation-turn-event.md)
- [CONVERSATION_ELICITATION_REPLY](../messages/conversation-elicitation-reply.md)

Capability implementation in [WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java) and [AgentConversationService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/AgentConversationService.java).
