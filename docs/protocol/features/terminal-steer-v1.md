# Steering a running dispatch

English | [中文](terminal-steer-v1.zh.md)

Heartbeat capability identifier: `terminal_steer_v1`.

Allows additional instructions during a dispatch. Return the outcome for the supplied `guidanceId`.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "terminal_steer_v1"
  ]
}
```

## Related protocol

- [TASK_STEER](../messages/task-steer.md)
- [TASK_STEER_ACK](../messages/task-steer-ack.md)

Capability implementation in [DispatchGuidanceService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DispatchGuidanceService.java).
