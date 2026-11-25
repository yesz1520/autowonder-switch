# Update and restart

English | [中文](executor-update-restart-v1.zh.md)

Heartbeat capability identifier: `EXECUTOR_UPDATE_RESTART_V1`.

Supports restart requests with `update: true`. Report update and restart progress through restart result frames.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "EXECUTOR_UPDATE_RESTART_V1"
  ]
}
```

## Related protocol

- [EXECUTOR_RESTART](../messages/executor-restart.md)

Capability implementation in [ExecutorRestartService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorRestartService.java).
