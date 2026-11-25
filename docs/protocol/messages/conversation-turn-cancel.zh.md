# 下发轮次取消

[English](conversation-turn-cancel.md) | 中文

平台 → 执行器。帧类型为 `CONVERSATION_TURN_CANCEL`。

按 `executorId`、`conversationId` 和 `turnId` 取消指定轮次。若提供 `dispatchAttempt`，只取消该次尝试。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `CONVERSATION_TURN_CANCEL`。 |
| `executorId` | 整数 |  | 该连接对应的执行器 ID。 |
| `conversationId` | 整数 |  | 接收或产生此消息的会话 ID。 |
| `turnId` | 整数 |  | 会话内的轮次 ID。 |
| `dispatchAttempt` | 整数 | 关联指定的轮次尝试时 | 该轮次的派发尝试编号，回复沿用下发值。 |

## 消息示例

```json
{
  "type": "CONVERSATION_TURN_CANCEL",
  "executorId": 42,
  "conversationId": 80,
  "turnId": 81,
  "dispatchAttempt": 1
}
```

## 对应回包

回包参数见 [CONVERSATION_TURN_ACK](conversation-turn-ack.zh.md)。

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

## 相关协议

- [CONVERSATION_TURN_ACK](conversation-turn-ack.zh.md)
- [CONVERSATION_TURN_CANCEL](../features/conversation-turn-cancel.zh.md)

消息处理见[WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java)。
