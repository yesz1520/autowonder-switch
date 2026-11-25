# 投递交互答复

[English](conversation-elicitation-reply.md) | 中文

平台 → 执行器。帧类型为 `CONVERSATION_ELICITATION_REPLY`。

用 `conversationId`、`turnId` 和 `requestId` 定位问题。读取 `action` 和可选的 `content`。V1 投递还带 `answerProtocolVersion`、`answerId`、`dispatchAttempt`、`remainingTtlMs`、`receiptOnly`。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `CONVERSATION_ELICITATION_REPLY`。 |
| `executorId` | 整数 |  | 该连接对应的执行器 ID。 |
| `conversationId` | 整数 |  | 接收或产生此消息的会话 ID。 |
| `turnId` | 整数 |  | 会话内的轮次 ID。 |
| `requestId` | 字符串 |  | 关联请求与后续消息的字符串 ID。 |
| `action` | 字符串 |  | 用户对交互问题的操作，如 accept。 |
| `content` | 对象 |  | 用户提交的答案对象，字段由原问题决定。 |
| `answerProtocolVersion` | 整数 | V1 答复投递 | 交互答复投递协议版本，V1 填 1。 |
| `answerId` | 字符串 |  | 本次答复的交付标识，回执沿用原值。 |
| `dispatchAttempt` | 整数 | 关联指定的轮次尝试时 | 该轮次的派发尝试编号，回复沿用下发值。 |
| `remainingTtlMs` | 整数 | V1 答复投递 | 答复剩余有效时间，单位为毫秒。 |
| `receiptOnly` | 布尔值 | V1 答复投递 | 是否仅请求已有投递的回执。 |

## 消息示例

```json
{
  "type": "CONVERSATION_ELICITATION_REPLY",
  "executorId": 42,
  "conversationId": 80,
  "turnId": 81,
  "requestId": "request-01",
  "action": "accept",
  "content": {
    "choice": "yes"
  },
  "answerProtocolVersion": 1,
  "answerId": "answer-01",
  "dispatchAttempt": 1,
  "remainingTtlMs": 30000,
  "receiptOnly": false
}
```

## 对应回包

回包参数见 [CONVERSATION_ELICITATION_REPLY_ACK](conversation-elicitation-reply-ack.zh.md)。

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

- [CONVERSATION_ELICITATION_REPLY_ACK](conversation-elicitation-reply-ack.zh.md)
- [CONVERSATION_ACP_INTERACTION_V1](../features/conversation-acp-interaction-v1.zh.md)
- [CONVERSATION_ANSWER_DELIVERY_V1](../features/conversation-answer-delivery-v1.zh.md)

消息处理见[WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java)。
