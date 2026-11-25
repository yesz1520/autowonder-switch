# 提交轮次最终回复

[English](conversation-turn-ack.md) | 中文

执行器 → 平台。帧类型为 `CONVERSATION_TURN_ACK`。

按 `conversationId`、`turnId` 和下发的 `dispatchAttempt` 返回最终状态。成功填 `SUCCESS` 并带 `replyMarkdown` 和 `sessionId`，失败填 `FAILED` 和 `error`，取消填 `CANCELED`。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `CONVERSATION_TURN_ACK`。 |
| `conversationId` | 整数 |  | 接收或产生此消息的会话 ID。 |
| `turnId` | 整数 |  | 会话内的轮次 ID。 |
| `dispatchAttempt` | 整数 | 关联指定的轮次尝试时 | 该轮次的派发尝试编号，回复沿用下发值。 |
| `status` | 字符串 | 轮次结束时 | 轮次结果，取 SUCCESS、FAILED 或 CANCELED。 |
| `replyMarkdown` | 字符串 | 有回复正文时 | 以 Markdown 表示的回复正文。 |
| `sessionId` | 字符串 |  | 执行本轮的原生会话 ID。 |
| `conversationDeliveryProtocolVersion` | 整数 | 启用轮次回执时 | 轮次事件与最终回复的回执协议版本。 |
| `usage` | 对象 | 上报模型用量时 | 本轮模型用量统计。 |
| `error` | 字符串 | 失败时 | 操作失败的原因。 |

## 消息示例

```json
{
  "type": "CONVERSATION_TURN_ACK",
  "conversationId": 80,
  "turnId": 81,
  "dispatchAttempt": 1,
  "status": "SUCCESS",
  "replyMarkdown": "Review complete",
  "sessionId": "session-1",
  "conversationDeliveryProtocolVersion": 1
}
```

## 对应回包

回包参数见 [CONVERSATION_DELIVERY_ACK](conversation-delivery-ack.zh.md)。

```json
{
  "type": "CONVERSATION_DELIVERY_ACK",
  "conversationDeliveryProtocolVersion": 1,
  "frameType": "CONVERSATION_TURN_ACK",
  "conversationId": 80,
  "turnId": 81,
  "dispatchAttempt": 1,
  "eventSeq": 0,
  "chunkIndex": 0
}
```

## 相关协议

- [会话轮次](../conversations.zh.md)
- [CONVERSATION_DELIVERY_ACK](conversation-delivery-ack.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
