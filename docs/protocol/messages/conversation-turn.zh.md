# 派发会话轮次

[English](conversation-turn.md) | 中文

平台 → 执行器。帧类型为 `CONVERSATION_TURN`。

接收 `conversationId` 和 `turnId` 指定的轮次。输入在 `content`，系统提示在 `systemPrompt`。有 `dispatchAttempt` 时，后续消息要关联本次尝试。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `CONVERSATION_TURN`。 |
| `executorId` | 整数 |  | 该连接对应的执行器 ID。 |
| `conversationId` | 整数 |  | 接收或产生此消息的会话 ID。 |
| `turnId` | 整数 |  | 会话内的轮次 ID。 |
| `agentId` | 整数 |  | 执行本次任务的智能体 ID。 |
| `dispatchAttempt` | 整数 | 关联指定的轮次尝试时 | 该轮次的派发尝试编号，回复沿用下发值。 |
| `content` | 字符串 |  | 向执行器投递的内容。 |
| `requestId` | 字符串 |  | 关联请求与后续消息的字符串 ID。 |
| `systemPrompt` | 字符串 |  | 本轮使用的系统提示。 |
| `cliSessionRef` | 字符串 |  | 需要续接的 CLI 会话引用，首次会话可为空字符串。 |
| `agentVersionId` | 整数 |  | 本次执行使用的智能体版本 ID。 |
| `capabilityDownloadUrl` | 字符串 |  | 本轮能力包的下载地址。 |
| `capabilitySha256` | 字符串 |  | 能力包的 SHA-256 摘要。 |
| `capabilityHash` | 字符串 |  | 本轮能力集合的标识摘要。 |
| `mcpToken` | 字符串 |  | 会话轮次访问 MCP 服务的令牌。 |
| `conversationDeliveryProtocolVersion` | 整数 | 启用轮次回执时 | 轮次事件与最终回复的回执协议版本。 |
| `conversationAnswerProtocolVersion` | 整数 | 启用答复回执时 | 轮次协商的交互答复协议版本。 |
| `mcpSecrets` | 对象 | 随本次配置 | 以不透明引用为键的本次 MCP 配置值。 |
| `environmentVariables` | 对象 | 随本次配置 | 本次执行的环境变量名与字符串值。 |
| `conversationActorProtocolVersion` | 整数 |  | 会话参与者协议的版本号。 |

## 消息示例

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

## 对应回包

回包参数见 [CONVERSATION_TURN_ACK](conversation-turn-ack.zh.md)。

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

## 相关协议

- [会话轮次](../conversations.zh.md)
- [CONVERSATION_TURN_ACK](conversation-turn-ack.zh.md)
- [AGENT_ENVIRONMENT_VARIABLES_V1](../features/agent-environment-variables-v1.zh.md)
- [ATTACHMENT_MANIFEST_V1](../features/attachment-manifest-v1.zh.md)
- [ARTIFACT_OUTPUT_V1](../features/artifact-output-v1.zh.md)
- [ACTION_PLAN_V1](../features/action-plan-v1.zh.md)

消息处理见[WsConversationTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsConversationTransport.java)。
