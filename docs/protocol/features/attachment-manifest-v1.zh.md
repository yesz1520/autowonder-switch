# 会话附件清单

[English](attachment-manifest-v1.md) | 中文

心跳能力标识为 `ATTACHMENT_MANIFEST_V1`。

声明会话执行器能在一次下发中接收用户选择的文件清单。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "ATTACHMENT_MANIFEST_V1"
  ]
}
```

## 相关协议

- [CONVERSATION_TURN](../messages/conversation-turn.zh.md)

能力定义与使用见[ConversationProtocolFeatures.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/ConversationProtocolFeatures.java)、[PlatformConversationService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/PlatformConversationService.java)。
