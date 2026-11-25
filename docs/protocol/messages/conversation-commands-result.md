# Return conversation commands

English | [中文](conversation-commands-result.zh.md)

Executor → platform. Frame type: `CONVERSATION_COMMANDS_RESULT`.

Echo `conversationId`. On success send `status: "OK"` and the serialized command list in `commands`. On failure describe the problem in `error`.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `CONVERSATION_COMMANDS_RESULT`. |
| `conversationId` | Integer |  | Conversation that receives or produces the message. |
| `status` | String |  | Use OK when the command list is available. |
| `commands` | String |  | Serialized JSON string with the list in availableCommands. |
| `error` | String | On failure | Description of the failure. |

## Example frame

```json
{
  "type": "CONVERSATION_COMMANDS_RESULT",
  "conversationId": 80,
  "status": "OK",
  "commands": "{\"availableCommands\":[]}"
}
```

## Related protocol

- [CONVERSATION_COMMANDS_PROBE](conversation-commands-probe.md)

Message handling in [ConversationCommandsService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/ConversationCommandsService.java) and [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
