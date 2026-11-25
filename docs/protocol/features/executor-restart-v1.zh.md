# 普通重启

[English](executor-restart-v1.md) | 中文

心跳能力标识为 `EXECUTOR_RESTART_V1`。

支持平台请求普通重启。新进程在心跳里用 `restartRequestId` 和 `startedAt` 确认启动。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "EXECUTOR_RESTART_V1"
  ]
}
```

## 相关协议

- [EXECUTOR_RESTART](../messages/executor-restart.zh.md)
- [EXECUTOR_RESTART_RESULT](../messages/executor-restart-result.zh.md)

能力定义与使用见[ExecutorRestartService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorRestartService.java)。
