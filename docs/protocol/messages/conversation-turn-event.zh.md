# 上报轮次事件

[English](conversation-turn-event.md) | 中文

执行器 → 平台。帧类型为 `CONVERSATION_TURN_EVENT`。

带回会话、轮次和尝试编号，用 `eventType`、`eventSeq`、从零开始的 `chunkIndex`、`chunkCount` 和 `payloadFragment` 上报事件。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `CONVERSATION_TURN_EVENT`。 |
| `conversationId` | 整数 |  | 接收或产生此消息的会话 ID。 |
| `turnId` | 整数 |  | 会话内的轮次 ID。 |
| `dispatchAttempt` | 整数 | 关联指定的轮次尝试时 | 该轮次的派发尝试编号，回复沿用下发值。 |
| `eventType` | 字符串 |  | 事件种类，决定如何解读事件内容。 |
| `eventSeq` | 整数 | 分片事件 | 轮次内的事件序号。 |
| `chunkIndex` | 整数 | 分片事件 | 分片下标，从 0 开始。 |
| `chunkCount` | 整数 | 分片事件 | 本事件的分片总数。 |
| `payloadFragment` | 字符串 | 分片事件 | 当前分片承载的序列化事件内容。 |
| `conversationDeliveryProtocolVersion` | 整数 | 启用轮次回执时 | 轮次事件与最终回复的回执协议版本。 |

## 消息示例

```json
{
  "type": "CONVERSATION_TURN_EVENT",
  "conversationId": 80,
  "turnId": 81,
  "dispatchAttempt": 1,
  "eventType": "status",
  "eventSeq": 1,
  "chunkIndex": 0,
  "chunkCount": 1,
  "payloadFragment": "{\"status\":\"running\"}",
  "conversationDeliveryProtocolVersion": 1
}
```

## 对应回包

回包参数见 [CONVERSATION_DELIVERY_ACK](conversation-delivery-ack.zh.md)。

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
- [CONVERSATION_DELIVERY_ACK](conversation-delivery-ack.zh.md)
- [CONVERSATION_TURN_EVENT](../features/conversation-turn-event.zh.md)
- [CONVERSATION_ACP_INTERACTION_V1](../features/conversation-acp-interaction-v1.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
