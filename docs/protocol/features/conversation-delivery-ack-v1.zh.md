# 轮次事件与结果回执

[English](conversation-delivery-ack-v1.md) | 中文

心跳能力标识为 `CONVERSATION_DELIVERY_ACK_V1`。

轮次带 `conversationDeliveryProtocolVersion: 1` 时，在事件和最终回复中带回版本，按回执匹配具体帧及分片。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "CONVERSATION_DELIVERY_ACK_V1"
  ]
}
```

## 相关协议

- [CONVERSATION_DELIVERY_ACK](../messages/conversation-delivery-ack.zh.md)

能力定义与使用见[WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java)、[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
