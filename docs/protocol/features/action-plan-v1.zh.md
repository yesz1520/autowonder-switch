# 行动计划确认

[English](action-plan-v1.md) | 中文

心跳能力标识为 `ACTION_PLAN_V1`。

平台管家会话通过会话令牌列出可用工具。需要确认的写操作通过 `autowonder.propose_platform_actions` 提交，参数 `actions` 包含 1 至 10 个 `{tool, arguments}` 对象。平台冻结参数并返回计划 ID、摘要和状态。执行器将待确认的计划告知用户；会话所有者在平台界面确认后，服务端按冻结的参数执行。

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
