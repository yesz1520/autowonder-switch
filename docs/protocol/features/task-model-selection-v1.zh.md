# 派单模型选择

[English](task-model-selection-v1.md) | 中文

心跳能力标识为 `TASK_MODEL_SELECTION_V1`。

按派单里的 `modelSelection.model` 选择本次任务使用的模型。心跳中的 `model` 用于上报当前模型。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "TASK_MODEL_SELECTION_V1"
  ]
}
```

## 相关协议

- [TASK_DISPATCH](../messages/task-dispatch.zh.md)

能力定义与使用见[WsDispatchTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchTransport.java)。
