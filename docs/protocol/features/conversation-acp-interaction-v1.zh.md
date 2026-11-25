# 轮次内的 ACP 交互

[English](conversation-acp-interaction-v1.md) | 中文

心跳能力标识为 `CONVERSATION_ACP_INTERACTION_V1`。

支持轮次中的 ACP 交互。执行器通过轮次事件发出提问，平台用同一轮次的答复消息投递用户选择。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "CONVERSATION_ACP_INTERACTION_V1"
  ]
}
```

## 相关协议

- [CONVERSATION_TURN_EVENT](../messages/conversation-turn-event.zh.md)
- [CONVERSATION_ELICITATION_REPLY](../messages/conversation-elicitation-reply.zh.md)

能力定义与使用见[WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java)、[AgentConversationService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/AgentConversationService.java)。
