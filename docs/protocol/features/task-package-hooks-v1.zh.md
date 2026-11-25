# 任务包 hook

[English](task-package-hooks-v1.md) | 中文

心跳能力标识为 `TASK_PACKAGE_HOOKS_V1`。

支持处理任务包声明的 hook。1.3.0 仍可能向未声明此能力的执行器发送含普通 hook 的包。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "TASK_PACKAGE_HOOKS_V1"
  ]
}
```

## 相关协议

- [接收派单](../tasks/dispatch.zh.md)

能力定义与使用见[WsDispatchTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchTransport.java)。
