# Confirm a pause

English | [中文](task-paused.zh.md)

Executor → platform. Frame type: `TASK_PAUSED`.

After uploading the recovery checkpoint, return `dispatchId`, `checkpointSeq` and `checkpointSha256` and wait for the platform receipt.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `TASK_PAUSED`. |
| `dispatchId` | Integer |  | ID of this dispatch attempt for the work item. |
| `checkpointSeq` | Integer | With a checkpoint | Positive checkpoint sequence number for the dispatch. |
| `checkpointSha256` | String | With a checkpoint | SHA-256 digest of the checkpoint. |

## Example frame

```json
{
  "type": "TASK_PAUSED",
  "dispatchId": 701,
  "checkpointSeq": 2,
  "checkpointSha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
```

## Example reply

See [TASK_RESULT_ACK](task-result-ack.md) for the reply fields.

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

## Related protocol

- [TASK_RESULT_ACK](task-result-ack.md)
- [Pause and recovery](../tasks/recovery.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
