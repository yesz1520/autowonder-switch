# 取消会话轮次

[English](conversation-turn-cancel.md) | 中文

心跳能力标识为 `CONVERSATION_TURN_CANCEL`。

支持按轮次取消。取消帧带 `dispatchAttempt` 时，只停止对应的派发尝试。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "CONVERSATION_TURN_CANCEL"
  ]
}
```

## 相关协议

- [CONVERSATION_TURN_CANCEL](../messages/conversation-turn-cancel.zh.md)

能力定义与使用见[WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java)。
