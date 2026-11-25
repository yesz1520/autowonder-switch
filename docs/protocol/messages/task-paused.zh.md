# 报告已暂停

[English](task-paused.md) | 中文

执行器 → 平台。帧类型为 `TASK_PAUSED`。

恢复检查点上传后，回传 `dispatchId`、`checkpointSeq` 和 `checkpointSha256`，等待平台确认。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `TASK_PAUSED`。 |
| `dispatchId` | 整数 |  | 本次派单 ID，同一工单重新派发时可能不同。 |
| `checkpointSeq` | 整数 | 关联检查点时 | 该派单的检查点序号，使用正整数。 |
| `checkpointSha256` | 字符串 | 关联检查点时 | 检查点的 SHA-256 摘要。 |

## 消息示例

```json
{
  "type": "TASK_PAUSED",
  "dispatchId": 701,
  "checkpointSeq": 2,
  "checkpointSha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
```

## 对应回包

回包参数见 [TASK_RESULT_ACK](task-result-ack.zh.md)。

```json
{
  "type": "TASK_RESULT_ACK",
  "dispatchId": 701,
  "accepted": true,
  "pauseCheckpointConfirmed": true,
  "checkpointSeq": 2,
  "checkpointSha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
```

## 相关协议

- [TASK_RESULT_ACK](task-result-ack.zh.md)
- [暂停与恢复](../tasks/recovery.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
