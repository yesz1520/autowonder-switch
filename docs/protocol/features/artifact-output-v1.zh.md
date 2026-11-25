# 会话产物输出

[English](artifact-output-v1.md) | 中文

心跳能力标识为 `ARTIFACT_OUTPUT_V1`。

声明会话执行器能把生成的文件作为产物交回平台，供用户预览或下载。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "ARTIFACT_OUTPUT_V1"
  ]
}
```

## 相关协议

- [CONVERSATION_TURN](../messages/conversation-turn.zh.md)

能力定义与使用见[ConversationProtocolFeatures.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/ConversationProtocolFeatures.java)、[PlatformConversationService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/PlatformConversationService.java)。
