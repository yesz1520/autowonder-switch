# Restart policy

English | [中文](executor-restart-policy-v1.zh.md)

Heartbeat capability identifier: `EXECUTOR_RESTART_POLICY_V1`.

Supports `restartMode`: `WHEN_IDLE` waits for idle and `FORCE` requests a forced restart. Advertise this alongside the applicable restart capability.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "EXECUTOR_RESTART_V1",
    "EXECUTOR_RESTART_POLICY_V1"
  ]
}
```

## Related protocol

- [EXECUTOR_RESTART](../messages/executor-restart.md)

Capability implementation in [ExecutorRestartService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorRestartService.java).
