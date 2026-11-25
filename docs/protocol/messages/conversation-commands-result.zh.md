# 返回会话命令

[English](conversation-commands-result.md) | 中文

执行器 → 平台。帧类型为 `CONVERSATION_COMMANDS_RESULT`。

带回 `conversationId`。成功时填 `status: "OK"` 和包含命令列表的 `commands` 字符串，失败时用 `error` 说明原因。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `CONVERSATION_COMMANDS_RESULT`。 |
| `conversationId` | 整数 |  | 接收或产生此消息的会话 ID。 |
| `status` | 字符串 |  | 命令列表可用时填 OK。 |
| `commands` | 字符串 |  | 序列化的 JSON 字符串，availableCommands 列出命令。 |
| `error` | 字符串 | 失败时 | 操作失败的原因。 |

## 消息示例

```json
{
  "type": "CONVERSATION_COMMANDS_RESULT",
  "conversationId": 80,
  "status": "OK",
  "commands": "{\"availableCommands\":[]}"
}
```

## 相关协议

- [CONVERSATION_COMMANDS_PROBE](conversation-commands-probe.zh.md)

消息处理见[ConversationCommandsService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/ConversationCommandsService.java)、[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
