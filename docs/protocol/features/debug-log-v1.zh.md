# 派单调试日志

[English](debug-log-v1.md) | 中文

心跳能力标识为 `DEBUG_LOG_V1`。

开启调试日志的派单可带 `debugLog.enabled` 和 `debugLog.maxBytes`。执行器按字节上限采集日志，通过结果的 `debugLog` 段上报。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "DEBUG_LOG_V1"
  ]
}
```

## 相关协议

- [TASK_DISPATCH](../messages/task-dispatch.zh.md)
- [TASK_RESULT](../messages/task-result.zh.md)

能力定义与使用见[WsDispatchTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchTransport.java)、[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
