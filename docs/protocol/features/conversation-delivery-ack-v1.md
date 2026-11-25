# Turn event and result receipts

English | [中文](conversation-delivery-ack-v1.zh.md)

Heartbeat capability identifier: `CONVERSATION_DELIVERY_ACK_V1`.

When a turn carries `conversationDeliveryProtocolVersion: 1`, echo it in events and the final reply. Match receipts to the individual frame and chunk.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "CONVERSATION_DELIVERY_ACK_V1"
  ]
}
```

## Related protocol

- [CONVERSATION_DELIVERY_ACK](../messages/conversation-delivery-ack.md)

Capability implementation in [WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java) and [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
