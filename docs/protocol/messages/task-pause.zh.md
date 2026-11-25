# 暂停或取消派单

[English](task-pause.md) | 中文

平台 → 执行器。帧类型为 `TASK_PAUSE`。

按 `executorId` 和 `dispatchId` 定位任务。普通暂停先保存恢复检查点，`cancel: true` 表示取消该派单。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `TASK_PAUSE`。 |
| `executorId` | 整数 |  | 该连接对应的执行器 ID。 |
| `dispatchId` | 整数 |  | 本次派单 ID，同一工单重新派发时可能不同。 |
| `cancel` | 布尔值 | 请求取消时 | 为 true 时取消派单，省略时请求暂停。 |

## 消息示例

```json
{
  "type": "TASK_PAUSE",
  "executorId": 42,
  "dispatchId": 701
}
```

## 对应回包

回包参数见 [TASK_PAUSED](task-paused.zh.md)。

```json
{
  "type": "TASK_PAUSED",
  "dispatchId": 701,
  "checkpointSeq": 2,
  "checkpointSha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
```

## 相关协议

- [TASK_PAUSED](task-paused.zh.md)
- [TASK_PAUSE_FAILED](task-pause-failed.zh.md)
- [暂停与恢复](../tasks/recovery.zh.md)

消息处理见[WsDispatchControlTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchControlTransport.java)。
