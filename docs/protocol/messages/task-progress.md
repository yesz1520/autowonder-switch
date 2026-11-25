# Report progress

English | [中文](task-progress.zh.md)

Executor → platform. Frame type: `TASK_PROGRESS`.

Identify the dispatch with `dispatchId` and describe progress with `eventType` and `message`. Include `stepOrder` or `stepName` when reporting a workflow step.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `TASK_PROGRESS`. |
| `dispatchId` | Integer |  | ID of this dispatch attempt for the work item. |
| `eventType` | String |  | Event kind used to interpret its payload. |
| `message` | String |  | Human-readable description of the operation. |
| `stepOrder` | Integer | Workflow step progress | Workflow step number. |
| `stepName` | String | Workflow step progress | Workflow step name. |

## Example frame

```json
{
  "type": "TASK_PROGRESS",
  "dispatchId": 701,
  "eventType": "package.completed",
  "message": "Package ready"
}
```

Version 1.3.0 defines no dedicated WebSocket receipt for this message.

## Structured events

Send the event kind in `eventType` and a description in `message`, either at the top level or inside `runtimeEvent`. The example uses `package.completed` to indicate that the task package is ready. Free-form text in `resultSummary` alone does not produce a structured progress event.

```json
{
  "type": "TASK_PROGRESS",
  "dispatchId": 701,
  "runtimeEvent": {
    "eventType": "package.completed",
    "eventId": "701-package",
    "seq": 1,
    "message": "Package ready"
  }
}
```

The `eventId` identifies a replay, `seq` is the event sequence number, and `message` is human-readable. [DispatchService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DispatchService.java)

## Related protocol

- [Accept a dispatch](../tasks/dispatch.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
