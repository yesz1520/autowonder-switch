# Lightweight checkpoint recovery

English | [中文](checkpoint-lightweight-v1.zh.md)

Heartbeat capability identifier: `CHECKPOINT_LIGHTWEIGHT_V1`.

Advertise support for reading lightweight checkpoints on resume.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "CHECKPOINT_LIGHTWEIGHT_V1"
  ]
}
```

## Related protocol

- [Pause and recovery](../tasks/recovery.md)

Capability implementation in [DispatchService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DispatchService.java).
