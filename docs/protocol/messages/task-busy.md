# Report a busy executor

English | [中文](task-busy.zh.md)

Executor → platform. Frame type: `TASK_BUSY`.

Return `dispatchId` when unable to accept the dispatch. The optional `reason` describes why.

## Message fields

| Field | JSON type | Description |
| --- | --- | --- |
| `type` | String | Fixed value `TASK_BUSY`. |
| `dispatchId` | Integer | ID of this dispatch attempt for the work item. |
| `reason` | String | Reason the executor cannot accept this dispatch. |

## Example frame

```json
{
  "type": "TASK_BUSY",
  "dispatchId": 701,
  "reason": "capacity reached"
}
```

Version 1.3.0 defines no dedicated WebSocket receipt for this message.

## Related protocol

- [TASK_DISPATCH](task-dispatch.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
