# Accept a dispatch

English | [中文](task-ack.zh.md)

Executor → platform. Frame type: `TASK_ACK`.

Echo the original `dispatchId` after accepting a dispatch. Report progress and the final result separately.

## Message fields

| Field | JSON type | Description |
| --- | --- | --- |
| `type` | String | Fixed value `TASK_ACK`. |
| `dispatchId` | Integer | ID of this dispatch attempt for the work item. |

## Example frame

```json
{
  "type": "TASK_ACK",
  "dispatchId": 701
}
```

Version 1.3.0 defines no dedicated WebSocket receipt for this message.

## Related protocol

- [TASK_DISPATCH](task-dispatch.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
