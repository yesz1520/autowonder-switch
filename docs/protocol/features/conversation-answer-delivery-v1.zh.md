# 交互答复投递

[English](conversation-answer-delivery-v1.md) | 中文

心跳能力标识为 `CONVERSATION_ANSWER_DELIVERY_V1`。

轮次可带 `conversationAnswerProtocolVersion: 1`。答复帧用 `answerProtocolVersion` 和 `answerId` 标识投递，执行器回传对应回执。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "CONVERSATION_ANSWER_DELIVERY_V1"
  ]
}
```

## 相关协议

- [CONVERSATION_ELICITATION_REPLY](../messages/conversation-elicitation-reply.zh.md)
- [CONVERSATION_ELICITATION_REPLY_ACK](../messages/conversation-elicitation-reply-ack.zh.md)

能力定义与使用见[WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java)。
