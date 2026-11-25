# 交付结果

[English](task-result.md) | 中文

执行器 → 平台。帧类型为 `TASK_RESULT`。

V2 结果携带 `dispatchId`、`resultProtocolVersion: 2`、`resultId` 和 `success`。成功时关联已上传的完成检查点，失败时填写 `error`。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `TASK_RESULT`。 |
| `dispatchId` | 整数 |  | 本次派单 ID，同一工单重新派发时可能不同。 |
| `resultProtocolVersion` | 整数 | V2 结果协议 | 结果协议版本，V2 填 2。 |
| `resultId` | 字符串 | V2 结果协议 | 本次结果的小写 UUID；重复交付沿用原值。 |
| `success` | 布尔值 |  | 操作是否成功。 |
| `error` | 字符串 | 失败时 | 操作失败的原因。 |
| `resultSummary` | 字符串 | 成功结果 | 任务结果摘要。 |
| `checkpointSeq` | 整数 | 成功结果 | 该派单的检查点序号，使用正整数。 |
| `checkpointSha256` | 字符串 | 成功结果 | 检查点的 SHA-256 摘要。 |
| `failureCategory` | 字符串 | 失败时 | 失败分类，随失败结果上报。 |
| `workflowChanged` | 布尔值 | 工作流发生变更时 | 本次执行是否改变了工作流。 |
| `handoff` | 对象 | 请求交接时 | 任务完成时建议交接的目标及类型。 |
| `workflowPlan` | 对象 | 工作流发生变更时 | 任务成功后要应用的工作流计划。 |
| `debugLog` | 对象 | 调试日志已开启 | 本次调试日志上传结果。 |

## 消息示例

```json
{
  "type": "TASK_RESULT",
  "dispatchId": 701,
  "resultProtocolVersion": 2,
  "resultId": "86ebc809-2d32-44e4-b456-29b870f878bd",
  "success": false,
  "error": "Model unavailable"
}
```

上传完成检查点后的成功结果：

```json
{
  "type": "TASK_RESULT",
  "dispatchId": 701,
  "resultProtocolVersion": 2,
  "resultId": "86ebc809-2d32-44e4-b456-29b870f878be",
  "success": true,
  "resultSummary": "Review complete",
  "checkpointSeq": 2,
  "checkpointSha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
```

## 对应回包

下例确认上面的失败报告已送达。回包参数见 [TASK_RESULT_ACK](task-result-ack.zh.md)。

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
- [TASK_RESULT_ACK](task-result-ack.zh.md)
- [DEBUG_LOG_V1](../features/debug-log-v1.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
