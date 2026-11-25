# Deliver a result

English | [中文](task-result.zh.md)

Executor → platform. Frame type: `TASK_RESULT`.

A V2 result carries `dispatchId`, `resultProtocolVersion: 2`, `resultId` and `success`. Reference the uploaded completion checkpoint on success and supply `error` on failure.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `TASK_RESULT`. |
| `dispatchId` | Integer |  | ID of this dispatch attempt for the work item. |
| `resultProtocolVersion` | Integer | V2 result protocol | Result protocol version. Use 2 for V2. |
| `resultId` | String | V2 result protocol | Lowercase UUID identifying the result across retries. |
| `success` | Boolean |  | Whether the operation succeeded. |
| `error` | String | On failure | Description of the failure. |
| `resultSummary` | String | Successful result | Summary of the task outcome. |
| `checkpointSeq` | Integer | Successful result | Positive checkpoint sequence number for the dispatch. |
| `checkpointSha256` | String | Successful result | SHA-256 digest of the checkpoint. |
| `failureCategory` | String | On failure | Classification of the reported failure. |
| `workflowChanged` | Boolean | When reporting workflow changes | Whether this execution changed the workflow. |
| `handoff` | Object | When proposing a handoff | Proposed handoff target and type at completion. |
| `workflowPlan` | Object | When reporting workflow changes | Workflow plan to apply after successful execution. |
| `debugLog` | Object | Debug logging enabled | Upload outcome for this dispatch’s debug log. |

## Example frame

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

A successful result after uploading the completion checkpoint:

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

## Example reply

This receipt confirms delivery of the failure report above. See [TASK_RESULT_ACK](task-result-ack.md) for the reply fields.

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

## Related protocol

- [Deliver results](../tasks/delivery.md)
- [TASK_RESULT_ACK](task-result-ack.md)
- [DEBUG_LOG_V1](../features/debug-log-v1.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
