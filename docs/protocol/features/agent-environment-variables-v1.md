# Execution environment variables

English | [中文](agent-environment-variables-v1.zh.md)

Heartbeat capability identifier: `AGENT_ENVIRONMENT_VARIABLES_V1`.

Accept `environmentVariables` on dispatches and conversation turns as a map of names to string values scoped to that execution.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "AGENT_ENVIRONMENT_VARIABLES_V1"
  ]
}
```

## Related protocol

- [TASK_DISPATCH](../messages/task-dispatch.md)
- [CONVERSATION_TURN](../messages/conversation-turn.md)

Capability implementation in [WsDispatchTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchTransport.java) and [WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java).
