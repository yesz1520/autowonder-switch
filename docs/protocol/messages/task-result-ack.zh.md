# 确认派单结果

[English](task-result-ack.md) | 中文

平台 → 执行器。帧类型为 `TASK_RESULT_ACK`。

平台用 `accepted` 表明是否接受报告。V2 结果按 `dispatchId` 和 `resultId` 匹配；暂停结果也使用此帧类型，可能携带检查点序号和摘要。

在收到与 `resultId` 匹配的结果回执前断线，重连后重发原结果。同一 ID 对应不同内容可能收到 `RESULT_ID_CONFLICT`。旧版回执和暂停回执均无法确认 V2 结果。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `TASK_RESULT_ACK`。 |
| `dispatchId` | 整数 |  | 本次派单 ID，同一工单重新派发时可能不同。 |
| `resultProtocolVersion` | 整数 | V2 结果协议 | 结果协议版本，V2 填 2。 |
| `resultId` | 字符串 | V2 结果协议 | 本次结果的小写 UUID；重复交付沿用原值。 |
| `accepted` | 布尔值 |  | 平台是否接收本次报告。 |
| `disposition` | 字符串 | V2 结果协议 | 结果处置状态，取 ACCEPTED、REJECTED 或 RETRYABLE。 |
| `retryable` | 布尔值 | V2 结果协议 | 此结果能否用相同 resultId 重试。 |
| `reasonCode` | 字符串 | 有原因代码时 | 服务端给出的原因代码。 |
| `pauseCheckpointConfirmed` | 布尔值 | 确认暂停检查点时 | 平台已确认暂停检查点时为 true。 |
| `checkpointSeq` | 整数 | 关联检查点时 | 该派单的检查点序号，使用正整数。 |
| `checkpointSha256` | 字符串 | 关联检查点时 | 检查点的 SHA-256 摘要。 |

## 消息示例

```json
{
  "type": "TASK_RESULT_ACK",
  "dispatchId": 701,
  "resultProtocolVersion": 2,
  "resultId": "86ebc809-2d32-44e4-b456-29b870f878bd",
  "accepted": true,
  "disposition": "ACCEPTED",
  "retryable": false
}
```

## 相关协议

- [交付结果](../tasks/delivery.zh.md)
- [TASK_PAUSED](task-paused.zh.md)
- [TASK_PAUSE_FAILED](task-pause-failed.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
