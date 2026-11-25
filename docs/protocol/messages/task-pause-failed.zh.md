# 报告暂停失败

[English](task-pause-failed.md) | 中文

执行器 → 平台。帧类型为 `TASK_PAUSE_FAILED`。

无法暂停或保存检查点时，回传 `dispatchId` 和 `error`。平台用结果回执确认收到报告。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `TASK_PAUSE_FAILED`。 |
| `dispatchId` | 整数 |  | 本次派单 ID，同一工单重新派发时可能不同。 |
| `error` | 字符串 | 失败时 | 操作失败的原因。 |

## 消息示例

```json
{
  "type": "TASK_PAUSE_FAILED",
  "dispatchId": 701,
  "error": "Model unavailable"
}
```

## 对应回包

回包参数见 [TASK_RESULT_ACK](task-result-ack.zh.md)。

```json
{
  "type": "TASK_RESULT_ACK",
  "dispatchId": 701,
  "accepted": true
}
```

## 相关协议

- [TASK_RESULT_ACK](task-result-ack.zh.md)
- [TASK_PAUSE](task-pause.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
