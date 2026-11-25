# Acknowledge task guidance

English | [中文](task-guidance-ack.zh.md)

Executor → platform. Frame type: `TASK_GUIDANCE_ACK`.

Echo `guidanceId` with a `status` of `APPLIED` or `FAILED`. Include `replyMarkdown` or `error` as appropriate. When submitting a `workflowPlan`, include its `dispatchId`.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `TASK_GUIDANCE_ACK`. |
| `guidanceId` | Integer |  | Guidance or steering instruction ID, echoed in the reply. |
| `dispatchId` | Integer |  | ID of this dispatch attempt for the work item. |
| `status` | String |  | Guidance outcome, APPLIED or FAILED. |
| `replyMarkdown` | String | When returning a reply | Reply body in Markdown. |
| `error` | String | On failure | Description of the failure. |
| `workflowPlan` | Object | When reporting workflow changes | Workflow plan produced by applied guidance. |

## Example frame

```json
{
  "type": "TASK_GUIDANCE_ACK",
  "guidanceId": 501,
  "dispatchId": 701,
  "status": "APPLIED",
  "replyMarkdown": "Review complete"
}
```

## Related protocol

- [TASK_GUIDANCE](task-guidance.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java) and [GuidanceService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/guidance/GuidanceService.java).
