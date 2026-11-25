# Pause or cancel a dispatch

English | [中文](task-pause.zh.md)

Platform → executor. Frame type: `TASK_PAUSE`.

Identify the task by `executorId` and `dispatchId`. Save a recovery checkpoint for an ordinary pause. A command with `cancel: true` cancels the dispatch.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `TASK_PAUSE`. |
| `executorId` | Integer |  | ID of the executor on this connection. |
| `dispatchId` | Integer |  | ID of this dispatch attempt for the work item. |
| `cancel` | Boolean | On cancellation | True cancels the dispatch. Omit to request a pause. |

## Example frame

```json
{
  "type": "TASK_PAUSE",
  "executorId": 42,
  "dispatchId": 701
}
```

## Example reply

See [TASK_PAUSED](task-paused.md) for the reply fields.

```json
{
  "type": "TASK_PAUSED",
  "dispatchId": 701,
  "checkpointSeq": 2,
  "checkpointSha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
```

## Related protocol

- [TASK_PAUSED](task-paused.md)
- [TASK_PAUSE_FAILED](task-pause-failed.md)
- [Pause and recovery](../tasks/recovery.md)

Message handling in [WsDispatchControlTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchControlTransport.java).
