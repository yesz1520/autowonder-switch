# Return an MCP test result

English | [中文](mcp-connection-test-result.zh.md)

Executor → platform. Frame type: `MCP_CONNECTION_TEST_RESULT`.

Echo `testId` and a boolean `success`. Report the test message, duration in milliseconds and discovered tools in `message`, `durationMs` and `tools`.

## Message fields

| Field | JSON type | Description |
| --- | --- | --- |
| `type` | String | Fixed value `MCP_CONNECTION_TEST_RESULT`. |
| `testId` | String | Correlation ID for the MCP connection test. |
| `success` | Boolean | Whether the operation succeeded. |
| `durationMs` | Integer | Elapsed test time in milliseconds. |
| `tools` | Array | MCP tools discovered during the test. |
| `message` | String | Human-readable description of the operation. |

## Example frame

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

- [MCP_CONNECTION_TEST](mcp-connection-test.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
