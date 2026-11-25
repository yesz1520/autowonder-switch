# Blocking tool hooks

English | [中文](task-package-tool-hooks-v1.zh.md)

Heartbeat capability identifier: `TASK_PACKAGE_TOOL_HOOKS_V1`.

Supports blocking tool hooks in a task package. An executor must advertise this feature to receive a package that requires them.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "TASK_PACKAGE_TOOL_HOOKS_V1"
  ]
}
```

## Related protocol

- [Accept a dispatch](../tasks/dispatch.md)

Capability implementation in [WsDispatchTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchTransport.java).
