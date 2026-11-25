# 阻断式工具 hook

[English](task-package-tool-hooks-v1.md) | 中文

心跳能力标识为 `TASK_PACKAGE_TOOL_HOOKS_V1`。

支持任务包中的阻断式工具 hook。任务包要求这种 hook 时，执行器必须声明此能力才能接收该包。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "TASK_PACKAGE_TOOL_HOOKS_V1"
  ]
}
```

## 相关协议

- [接收派单](../tasks/dispatch.zh.md)

能力定义与使用见[WsDispatchTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchTransport.java)。
