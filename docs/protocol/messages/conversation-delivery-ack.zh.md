# 确认收到轮次消息

[English](conversation-delivery-ack.md) | 中文

平台 → 执行器。帧类型为 `CONVERSATION_DELIVERY_ACK`。

用 `conversationDeliveryProtocolVersion: 1`、`frameType`、`conversationId`、`turnId` 和 `dispatchAttempt` 匹配原帧。事件还要按 `eventSeq` 和 `chunkIndex` 匹配。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `CONVERSATION_DELIVERY_ACK`。 |
| `conversationDeliveryProtocolVersion` | 整数 | 启用轮次回执时 | 轮次事件与最终回复的回执协议版本。 |
| `frameType` | 字符串 |  | 本回执确认的原始帧类型。 |
| `conversationId` | 整数 |  | 接收或产生此消息的会话 ID。 |
| `turnId` | 整数 |  | 会话内的轮次 ID。 |
| `dispatchAttempt` | 整数 | 关联指定的轮次尝试时 | 该轮次的派发尝试编号，回复沿用下发值。 |
| `eventSeq` | 整数 | 每次回执 | 轮次内的事件序号。 确认最终回复时通常为 0。 |
| `chunkIndex` | 整数 | 每次回执 | 分片下标，从 0 开始。 确认最终回复时通常为 0。 |

## 消息示例

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

## 相关协议

- [会话轮次](../conversations.zh.md)
- [CONVERSATION_DELIVERY_ACK_V1](../features/conversation-delivery-ack-v1.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
