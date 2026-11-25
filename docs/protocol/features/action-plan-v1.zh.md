# 行动计划确认

[English](action-plan-v1.md) | 中文

心跳能力标识为 `ACTION_PLAN_V1`。

声明可接收参数已冻结的行动计划，并在执行前等待用户确认。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "ACTION_PLAN_V1"
  ]
}
```

## 相关协议

- [CONVERSATION_TURN](../messages/conversation-turn.zh.md)

能力定义与使用见[ConversationProtocolFeatures.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/ConversationProtocolFeatures.java)、[PlatformConversationService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/PlatformConversationService.java)。
