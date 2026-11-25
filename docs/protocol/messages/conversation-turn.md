# Dispatch a turn

English | [中文](conversation-turn.zh.md)

Platform → executor. Frame type: `CONVERSATION_TURN`.

Receive the turn identified by `conversationId` and `turnId`. Read user input from `content` and the system prompt from `systemPrompt`. Preserve `dispatchAttempt` when supplied.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `CONVERSATION_TURN`. |
| `executorId` | Integer |  | ID of the executor on this connection. |
| `conversationId` | Integer |  | Conversation that receives or produces the message. |
| `turnId` | Integer |  | Turn ID within the conversation. |
| `agentId` | Integer |  | Agent assigned to this execution. |
| `dispatchAttempt` | Integer | For the specified turn attempt | Turn delivery attempt number. Echo the supplied value. |
| `content` | String |  | Content delivered to the executor. |
| `requestId` | String |  | String identifier correlating the request with subsequent messages. |
| `systemPrompt` | String |  | System instructions for the turn. |
| `cliSessionRef` | String |  | CLI session to resume. May be empty for a new session. |
| `agentVersionId` | Integer |  | Agent version used for this execution. |
| `capabilityDownloadUrl` | String |  | Download URL for the turn’s capability archive. |
| `capabilitySha256` | String |  | SHA-256 digest of the capability archive. |
| `capabilityHash` | String |  | Hash identifying the capability set for this turn. |
| `mcpToken` | String |  | Token for MCP calls during this turn. |
| `conversationDeliveryProtocolVersion` | Integer | Turn receipts enabled | Receipt protocol version for turn events and final replies. |
| `conversationAnswerProtocolVersion` | Integer | Answer receipts enabled | Interactive answer protocol version negotiated for the turn. |
| `mcpSecrets` | Object | When configured | Per-execution MCP configuration values keyed by opaque references. |
| `environmentVariables` | Object | When configured | Environment variable names and string values for this execution. |
| `conversationActorProtocolVersion` | Integer |  | Conversation actor protocol version. |

## Example frame

```json
{
  "type": "CONVERSATION_TURN",
  "executorId": 42,
  "conversationId": 80,
  "turnId": 81,
  "agentId": 7,
  "dispatchAttempt": 1,
  "content": "Review the changes",
  "requestId": "request-01",
  "systemPrompt": "Review the patch.",
  "cliSessionRef": "session-1",
  "agentVersionId": 11,
  "capabilityDownloadUrl": "https://storage.example/capabilities.zip",
  "capabilitySha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "capabilityHash": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
  "mcpToken": "EXAMPLE",
  "conversationDeliveryProtocolVersion": 1
}
```

## Example reply

See [CONVERSATION_TURN_ACK](conversation-turn-ack.md) for the reply fields.

```json
{
  "type": "CONVERSATION_TURN_ACK",
  "conversationId": 80,
  "turnId": 81,
  "dispatchAttempt": 1,
  "status": "SUCCESS",
  "replyMarkdown": "Review complete",
  "sessionId": "session-1",
  "conversationDeliveryProtocolVersion": 1
}
```

## Related protocol

- [Conversation turns](../conversations.md)
- [CONVERSATION_TURN_ACK](conversation-turn-ack.md)
- [AGENT_ENVIRONMENT_VARIABLES_V1](../features/agent-environment-variables-v1.md)
- [ATTACHMENT_MANIFEST_V1](../features/attachment-manifest-v1.md)
- [ARTIFACT_OUTPUT_V1](../features/artifact-output-v1.md)
- [ACTION_PLAN_V1](../features/action-plan-v1.md)

Message handling in [WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java).
