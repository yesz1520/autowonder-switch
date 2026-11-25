# 返回 MCP 测试结果

[English](mcp-connection-test-result.md) | 中文

执行器 → 平台。帧类型为 `MCP_CONNECTION_TEST_RESULT`。

回传 `testId` 和布尔值 `success`，用 `message`、`durationMs` 和 `tools` 上报测试说明、毫秒耗时和工具列表。

## 消息参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `type` | 字符串 | 固定为 `MCP_CONNECTION_TEST_RESULT`。 |
| `testId` | 字符串 | MCP 连接测试的关联 ID。 |
| `success` | 布尔值 | 操作是否成功。 |
| `durationMs` | 整数 | 连接测试耗时，单位为毫秒。 |
| `tools` | 数组 | 测试时发现的 MCP 工具列表。 |
| `message` | 字符串 | 本次操作的可读说明。 |

## 消息示例

```json
{
  "type": "MCP_CONNECTION_TEST_RESULT",
  "testId": "test-01",
  "success": true,
  "durationMs": 24,
  "tools": []
}
```

## 相关协议

- [MCP_CONNECTION_TEST](mcp-connection-test.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
