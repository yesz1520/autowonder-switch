# 确认交互答复

[English](conversation-elicitation-reply-ack.md) | 中文

执行器 → 平台。帧类型为 `CONVERSATION_ELICITATION_REPLY_ACK`。

回传 `answerProtocolVersion: 1`、`conversationId`、`turnId`、`dispatchAttempt`、`requestId` 和 `answerId`。投递状态 `status` 可为 `RECEIVED`、`WRITTEN`、`UNAVAILABLE` 或 `UNKNOWN`。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `CONVERSATION_ELICITATION_REPLY_ACK`。 |
| `conversationId` | 整数 |  | 接收或产生此消息的会话 ID。 |
| `turnId` | 整数 |  | 会话内的轮次 ID。 |
| `requestId` | 字符串 |  | 关联请求与后续消息的字符串 ID。 |
| `answerId` | 字符串 |  | 本次答复的交付标识，回执沿用原值。 |
| `dispatchAttempt` | 整数 | 关联指定的轮次尝试时 | 该轮次的派发尝试编号，回复沿用下发值。 |
| `answerProtocolVersion` | 整数 | V1 答复投递 | 交互答复投递协议版本，V1 填 1。 |
| `status` | 字符串 |  | 答复的投递阶段：RECEIVED、WRITTEN、UNAVAILABLE 或 UNKNOWN。 |

## 消息示例

```json
{
  "type": "CONVERSATION_ELICITATION_REPLY_ACK",
  "conversationId": 80,
  "turnId": 81,
  "requestId": "request-01",
  "answerId": "answer-01",
  "dispatchAttempt": 1,
  "answerProtocolVersion": 1,
  "status": "WRITTEN"
}
```

## 相关协议

- [CONVERSATION_ELICITATION_REPLY](conversation-elicitation-reply.zh.md)
- [CONVERSATION_ANSWER_DELIVERY_V1](../features/conversation-answer-delivery-v1.zh.md)

消息处理见[ConversationAnswerDeliveryService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/ConversationAnswerDeliveryService.java)。
