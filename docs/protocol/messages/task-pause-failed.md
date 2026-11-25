# Report a failed pause

English | [中文](task-pause-failed.zh.md)

Executor → platform. Frame type: `TASK_PAUSE_FAILED`.

Return `dispatchId` and `error` if pausing or saving the checkpoint fails. The platform acknowledges the report with a result receipt.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `TASK_PAUSE_FAILED`. |
| `dispatchId` | Integer |  | ID of this dispatch attempt for the work item. |
| `error` | String | On failure | Description of the failure. |

## Example frame

```json
{
  "type": "TASK_PAUSE_FAILED",
  "dispatchId": 701,
  "error": "Model unavailable"
}
```

## Example reply

See [TASK_RESULT_ACK](task-result-ack.md) for the reply fields.

```json
{
  "type": "TASK_RESULT_ACK",
  "dispatchId": 701,
  "accepted": true
}
```

## Related protocol

- [TASK_RESULT_ACK](task-result-ack.md)
- [TASK_PAUSE](task-pause.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
