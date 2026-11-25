# 查询会话命令

[English](conversation-commands-probe.md) | 中文

平台 → 执行器。帧类型为 `CONVERSATION_COMMANDS_PROBE`。

平台按 `executorId` 和 `conversationId` 查询当前会话可用的 CLI 命令。

## 消息参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `type` | 字符串 | 固定为 `CONVERSATION_COMMANDS_PROBE`。 |
| `executorId` | 整数 | 该连接对应的执行器 ID。 |
| `conversationId` | 整数 | 接收或产生此消息的会话 ID。 |

## 消息示例

```json
{
  "type": "CONVERSATION_COMMANDS_PROBE",
  "executorId": 42,
  "conversationId": 80
}
```

## 对应回包

回包参数见 [CONVERSATION_COMMANDS_RESULT](conversation-commands-result.zh.md)。

```json
{
  "type": "CONVERSATION_COMMANDS_RESULT",
  "conversationId": 80,
  "status": "OK",
  "commands": "{\"availableCommands\":[]}"
}
```

## 相关协议

- [CONVERSATION_COMMANDS_RESULT](conversation-commands-result.zh.md)

消息处理见[WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java)。
