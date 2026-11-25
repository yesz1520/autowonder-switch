# Per-dispatch model selection

English | [中文](task-model-selection-v1.zh.md)

Heartbeat capability identifier: `TASK_MODEL_SELECTION_V1`.

Use `modelSelection.model` from the dispatch for this task. The heartbeat’s `model` reports the current model.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "TASK_MODEL_SELECTION_V1"
  ]
}
```

## Related protocol

- [TASK_DISPATCH](../messages/task-dispatch.md)

Capability implementation in [WsDispatchTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchTransport.java).
