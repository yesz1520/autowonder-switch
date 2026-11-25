# 轮次事件

[English](conversation-turn-event.md) | 中文

心跳能力标识为 `CONVERSATION_TURN_EVENT`。

在会话轮次结束前上报事件。每个事件用序号标识，大负载可以分片。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "CONVERSATION_TURN_EVENT"
  ]
}
```

## 相关协议

- [CONVERSATION_TURN_EVENT](../messages/conversation-turn-event.zh.md)

能力定义与使用见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
