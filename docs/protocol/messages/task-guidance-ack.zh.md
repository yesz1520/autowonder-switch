# 回复任务指导

[English](task-guidance-ack.md) | 中文

执行器 → 平台。帧类型为 `TASK_GUIDANCE_ACK`。

带回 `guidanceId`，`status` 为 `APPLIED` 或 `FAILED`。可附 `replyMarkdown` 或 `error`。提交 `workflowPlan` 时，还用 `dispatchId` 关联派单。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `TASK_GUIDANCE_ACK`。 |
| `guidanceId` | 整数 |  | 指导或追加指令的 ID，回复时原样带回。 |
| `dispatchId` | 整数 |  | 本次派单 ID，同一工单重新派发时可能不同。 |
| `status` | 字符串 |  | 指导是否已应用，取 APPLIED 或 FAILED。 |
| `replyMarkdown` | 字符串 | 有回复正文时 | 以 Markdown 表示的回复正文。 |
| `error` | 字符串 | 失败时 | 操作失败的原因。 |
| `workflowPlan` | 对象 | 工作流发生变更时 | 已应用的指导所产生的工作流计划。 |

## 消息示例

```json
{
  "type": "TASK_GUIDANCE_ACK",
  "guidanceId": 501,
  "dispatchId": 701,
  "status": "APPLIED",
  "replyMarkdown": "Review complete"
}
```

## 相关协议

- [TASK_GUIDANCE](task-guidance.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)、[GuidanceService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/guidance/GuidanceService.java)。
