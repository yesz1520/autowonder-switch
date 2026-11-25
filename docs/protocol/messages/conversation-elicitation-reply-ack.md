# Acknowledge an interactive answer

English | [中文](conversation-elicitation-reply-ack.zh.md)

Executor → platform. Frame type: `CONVERSATION_ELICITATION_REPLY_ACK`.

Echo `answerProtocolVersion: 1`, `conversationId`, `turnId`, `dispatchAttempt`, `requestId` and `answerId`. Delivery `status` is one of `RECEIVED`, `WRITTEN`, `UNAVAILABLE` or `UNKNOWN`.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `CONVERSATION_ELICITATION_REPLY_ACK`. |
| `conversationId` | Integer |  | Conversation that receives or produces the message. |
| `turnId` | Integer |  | Turn ID within the conversation. |
| `requestId` | String |  | String identifier correlating the request with subsequent messages. |
| `answerId` | String |  | Delivery ID of this answer. Echo it in the receipt. |
| `dispatchAttempt` | Integer | For the specified turn attempt | Turn delivery attempt number. Echo the supplied value. |
| `answerProtocolVersion` | Integer | V1 answer delivery | Version of the answer delivery protocol. Use 1 for V1. |
| `status` | String |  | Delivery stage: RECEIVED, WRITTEN, UNAVAILABLE or UNKNOWN. |

## Example frame

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

## Related protocol

- [CONVERSATION_ELICITATION_REPLY](conversation-elicitation-reply.md)
- [CONVERSATION_ANSWER_DELIVERY_V1](../features/conversation-answer-delivery-v1.md)

Message handling in [ConversationAnswerDeliveryService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/ConversationAnswerDeliveryService.java).
