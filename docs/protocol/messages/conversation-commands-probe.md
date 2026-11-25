# Probe conversation commands

English | [中文](conversation-commands-probe.zh.md)

Platform → executor. Frame type: `CONVERSATION_COMMANDS_PROBE`.

The platform asks which CLI commands are available for the given `executorId` and `conversationId`.

## Message fields

| Field | JSON type | Description |
| --- | --- | --- |
| `type` | String | Fixed value `CONVERSATION_COMMANDS_PROBE`. |
| `executorId` | Integer | ID of the executor on this connection. |
| `conversationId` | Integer | Conversation that receives or produces the message. |

## Example frame

```json
{
  "type": "CONVERSATION_COMMANDS_PROBE",
  "executorId": 42,
  "conversationId": 80
}
```

## Example reply

See [CONVERSATION_COMMANDS_RESULT](conversation-commands-result.md) for the reply fields.

```json
{
  "type": "CONVERSATION_COMMANDS_RESULT",
  "conversationId": 80,
  "status": "OK",
  "commands": "{\"availableCommands\":[]}"
}
```

## Related protocol

- [CONVERSATION_COMMANDS_RESULT](conversation-commands-result.md)

Message handling in [WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java).
