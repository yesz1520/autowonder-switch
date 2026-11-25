# Restart with launch configuration

English | [中文](executor-launch-command-v1.zh.md)

Heartbeat capability identifier: `EXECUTOR_LAUNCH_COMMAND_V1`.

Supports restart commands with a `launch` object. The command carries `launchConfigVersion` and can include the configured `targetRuntimeVersion`.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "EXECUTOR_LAUNCH_COMMAND_V1"
  ]
}
```

## Related protocol

- [EXECUTOR_RESTART](../messages/executor-restart.md)

Capability implementation in [ExecutorRestartService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorRestartService.java).
