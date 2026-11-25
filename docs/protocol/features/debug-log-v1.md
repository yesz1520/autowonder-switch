# Dispatch debug logs

English | [中文](debug-log-v1.zh.md)

Heartbeat capability identifier: `DEBUG_LOG_V1`.

Dispatches with debug logging enabled can carry `debugLog.enabled` and `debugLog.maxBytes`. Honor the byte limit and report log metadata in the result’s `debugLog` object.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "DEBUG_LOG_V1"
  ]
}
```

## Related protocol

- [TASK_DISPATCH](../messages/task-dispatch.md)
- [TASK_RESULT](../messages/task-result.md)

Capability implementation in [WsDispatchTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchTransport.java) and [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
