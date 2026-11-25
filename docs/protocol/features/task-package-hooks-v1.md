# Task package hooks

English | [中文](task-package-hooks-v1.zh.md)

Heartbeat capability identifier: `TASK_PACKAGE_HOOKS_V1`.

Supports hooks declared in the task package. In 1.3.0 an executor can still receive a package with ordinary hooks without advertising this feature.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "TASK_PACKAGE_HOOKS_V1"
  ]
}
```

## Related protocol

- [Accept a dispatch](../tasks/dispatch.md)

Capability implementation in [WsDispatchTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchTransport.java).
