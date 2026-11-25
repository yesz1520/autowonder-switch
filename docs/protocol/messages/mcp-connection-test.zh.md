# 测试 MCP 连接

[English](mcp-connection-test.md) | 中文

平台 → 执行器。帧类型为 `MCP_CONNECTION_TEST`。

用 `testId` 标识测试，按 `transport` 读取连接参数。命令传输使用 `command`、`args`、`env`；网络传输使用 `url`、`headers`。超时单位是秒（`timeoutSeconds`）。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `MCP_CONNECTION_TEST`。 |
| `executorId` | 整数 |  | 该连接对应的执行器 ID。 |
| `testId` | 字符串 |  | MCP 连接测试的关联 ID。 |
| `transport` | 字符串 |  | MCP 使用的连接方式。 |
| `command` | 字符串 | stdio 传输 | stdio 传输启动的程序。 |
| `args` | 数组 | stdio 传输 | 传给启动程序的字符串参数数组。 |
| `env` | 对象 | stdio 传输 | 程序启动时使用的环境变量。 |
| `timeoutSeconds` | 整数 |  | 连接测试最长等待秒数。 |
| `url` | 字符串 | 网络传输 | 网络传输的 MCP 服务地址。 |
| `headers` | 对象 | 网络传输 | 网络传输使用的请求头。 |

## 消息示例

```json
{
  "type": "MCP_CONNECTION_TEST",
  "executorId": 42,
  "testId": "test-01",
  "transport": "stdio",
  "command": "example-mcp",
  "args": [],
  "env": {},
  "timeoutSeconds": 60
}
```

## 对应回包

回包参数见 [MCP_CONNECTION_TEST_RESULT](mcp-connection-test-result.zh.md)。

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

- [MCP_CONNECTION_TEST_RESULT](mcp-connection-test-result.zh.md)

消息处理见[RuntimeMcpConnectionTestService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/skill/RuntimeMcpConnectionTestService.java)。
