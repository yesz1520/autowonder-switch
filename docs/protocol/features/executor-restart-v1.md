# Restart

English | [中文](executor-restart-v1.zh.md)

Heartbeat capability identifier: `EXECUTOR_RESTART_V1`.

Supports a platform-requested restart. The new process reports `restartRequestId` and `startedAt` in its heartbeat to confirm startup.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "EXECUTOR_RESTART_V1"
  ]
}
```

## Related protocol

- [EXECUTOR_RESTART](../messages/executor-restart.md)
- [EXECUTOR_RESTART_RESULT](../messages/executor-restart-result.md)

Capability implementation in [ExecutorRestartService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorRestartService.java).
