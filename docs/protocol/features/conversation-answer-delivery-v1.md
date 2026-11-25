# Interactive answer delivery

English | [中文](conversation-answer-delivery-v1.zh.md)

Heartbeat capability identifier: `CONVERSATION_ANSWER_DELIVERY_V1`.

Turns can carry `conversationAnswerProtocolVersion: 1`. Answer deliveries use `answerProtocolVersion` and `answerId`. return the matching receipt.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "CONVERSATION_ANSWER_DELIVERY_V1"
  ]
}
```

## Related protocol

- [CONVERSATION_ELICITATION_REPLY](../messages/conversation-elicitation-reply.md)
- [CONVERSATION_ELICITATION_REPLY_ACK](../messages/conversation-elicitation-reply-ack.md)

Capability implementation in [WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java).
