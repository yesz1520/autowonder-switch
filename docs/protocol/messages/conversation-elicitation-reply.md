# Deliver an interactive answer

English | [中文](conversation-elicitation-reply.zh.md)

Platform → executor. Frame type: `CONVERSATION_ELICITATION_REPLY`.

Identify the question by `conversationId`, `turnId` and `requestId`. Read `action` and optional `content`. V1 delivery adds `answerProtocolVersion`, `answerId`, `dispatchAttempt`, `remainingTtlMs` and `receiptOnly`.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `CONVERSATION_ELICITATION_REPLY`. |
| `executorId` | Integer |  | ID of the executor on this connection. |
| `conversationId` | Integer |  | Conversation that receives or produces the message. |
| `turnId` | Integer |  | Turn ID within the conversation. |
| `requestId` | String |  | String identifier correlating the request with subsequent messages. |
| `action` | String |  | User’s action on the interactive question, such as accept. |
| `content` | Object |  | Answer object with fields determined by the original question. |
| `answerProtocolVersion` | Integer | V1 answer delivery | Version of the answer delivery protocol. Use 1 for V1. |
| `answerId` | String |  | Delivery ID of this answer. Echo it in the receipt. |
| `dispatchAttempt` | Integer | For the specified turn attempt | Turn delivery attempt number. Echo the supplied value. |
| `remainingTtlMs` | Integer | V1 answer delivery | Remaining lifetime of the answer in milliseconds. |
| `receiptOnly` | Boolean | V1 answer delivery | Whether the platform is only requesting a receipt for an existing delivery. |

## Example frame

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

## Example reply

See [CONVERSATION_ELICITATION_REPLY_ACK](conversation-elicitation-reply-ack.md) for the reply fields.

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

- [CONVERSATION_ELICITATION_REPLY_ACK](conversation-elicitation-reply-ack.md)
- [CONVERSATION_ACP_INTERACTION_V1](../features/conversation-acp-interaction-v1.md)
- [CONVERSATION_ANSWER_DELIVERY_V1](../features/conversation-answer-delivery-v1.md)

Message handling in [WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java).
