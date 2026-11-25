# Human requests during dispatches

English | [中文](dispatch-requests-v1.zh.md)

Heartbeat capability identifier: `DISPATCH_REQUESTS_V1`.

Advertise support for dispatches that require human requests.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "DISPATCH_REQUESTS_V1"
  ]
}
```

## Related protocol

- [TASK_DISPATCH](../messages/task-dispatch.md)

Capability implementation in [DispatchService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DispatchService.java).
