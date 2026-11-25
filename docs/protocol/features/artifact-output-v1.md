# Conversation output artifacts

English | [中文](artifact-output-v1.zh.md)

Heartbeat capability identifier: `ARTIFACT_OUTPUT_V1`.

Advertise support for returning files generated in a conversation as artifacts for preview or download.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "ARTIFACT_OUTPUT_V1"
  ]
}
```

## Related protocol

- [CONVERSATION_TURN](../messages/conversation-turn.md)

Capability implementation in [ConversationProtocolFeatures.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/ConversationProtocolFeatures.java) and [PlatformConversationService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/PlatformConversationService.java).
