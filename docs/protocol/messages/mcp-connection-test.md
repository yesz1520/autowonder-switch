# Test an MCP connection

English | [中文](mcp-connection-test.zh.md)

Platform → executor. Frame type: `MCP_CONNECTION_TEST`.

Identify the test by `testId`. Read connection parameters for the specified `transport`: `command`, `args` and `env` for a command, or `url` and `headers` for network transport. The timeout is `timeoutSeconds`.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `MCP_CONNECTION_TEST`. |
| `executorId` | Integer |  | ID of the executor on this connection. |
| `testId` | String |  | Correlation ID for the MCP connection test. |
| `transport` | String |  | Transport used to connect to the MCP server. |
| `command` | String | stdio transport | Program launched for stdio transport. |
| `args` | Array | stdio transport | String arguments passed to the command. |
| `env` | Object | stdio transport | Environment variables supplied to the command. |
| `timeoutSeconds` | Integer |  | Connection test timeout in seconds. |
| `url` | String | Network transport | MCP server URL for network transports. |
| `headers` | Object | Network transport | Request headers for network transports. |

## Example frame

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

## Example reply

See [MCP_CONNECTION_TEST_RESULT](mcp-connection-test-result.md) for the reply fields.

```json
{
  "type": "MCP_CONNECTION_TEST_RESULT",
  "testId": "test-01",
  "success": true,
  "durationMs": 24,
  "tools": []
}
```

## Related protocol

- [MCP_CONNECTION_TEST_RESULT](mcp-connection-test-result.md)

Message handling in [RuntimeMcpConnectionTestService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/skill/RuntimeMcpConnectionTestService.java).
