# 执行环境变量

[English](agent-environment-variables-v1.md) | 中文

心跳能力标识为 `AGENT_ENVIRONMENT_VARIABLES_V1`。

接收派单或会话轮次里的 `environmentVariables`，它是本次执行的变量名到字符串值的映射。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "AGENT_ENVIRONMENT_VARIABLES_V1"
  ]
}
```

## 相关协议

- [TASK_DISPATCH](../messages/task-dispatch.zh.md)
- [CONVERSATION_TURN](../messages/conversation-turn.zh.md)

能力定义与使用见[WsDispatchTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchTransport.java)、[WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java)。
