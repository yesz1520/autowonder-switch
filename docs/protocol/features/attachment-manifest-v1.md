# Conversation attachment manifests

English | [中文](attachment-manifest-v1.zh.md)

Heartbeat capability identifier: `ATTACHMENT_MANIFEST_V1`.

Advertise support for receiving the manifest of files selected by the user in one delivery.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "ATTACHMENT_MANIFEST_V1"
  ]
}
```

## Related protocol

- [CONVERSATION_TURN](../messages/conversation-turn.md)

Capability implementation in [ConversationProtocolFeatures.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/ConversationProtocolFeatures.java) and [PlatformConversationService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/PlatformConversationService.java).
