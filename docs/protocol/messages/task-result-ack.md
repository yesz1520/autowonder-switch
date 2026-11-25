# Acknowledge a dispatch outcome

English | [中文](task-result-ack.zh.md)

Platform → executor. Frame type: `TASK_RESULT_ACK`.

The platform indicates acceptance with `accepted`. Match V2 results by `dispatchId` and `resultId`. Pause reports also use this receipt type and can carry the checkpoint sequence and digest.

If the connection closes before a matching result receipt arrives, resend the original result after reconnecting. Reusing a result ID for different content can yield `RESULT_ID_CONFLICT`. Legacy and pause receipts cannot confirm a V2 result.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `TASK_RESULT_ACK`. |
| `dispatchId` | Integer |  | ID of this dispatch attempt for the work item. |
| `resultProtocolVersion` | Integer | V2 result protocol | Result protocol version. Use 2 for V2. |
| `resultId` | String | V2 result protocol | Lowercase UUID identifying the result across retries. |
| `accepted` | Boolean |  | Whether the platform accepted the report. |
| `disposition` | String | V2 result protocol | Result disposition: ACCEPTED, REJECTED or RETRYABLE. |
| `retryable` | Boolean | V2 result protocol | Whether to retry with the same resultId. |
| `reasonCode` | String | When a reason code is available | Machine-readable reason for the outcome. |
| `pauseCheckpointConfirmed` | Boolean | On pause checkpoint confirmation | True when the platform confirms the pause checkpoint. |
| `checkpointSeq` | Integer | With a checkpoint | Positive checkpoint sequence number for the dispatch. |
| `checkpointSha256` | String | With a checkpoint | SHA-256 digest of the checkpoint. |

## Example frame

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
- [TASK_PAUSED](task-paused.md)
- [TASK_PAUSE_FAILED](task-pause-failed.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
