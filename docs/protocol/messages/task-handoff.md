# Request a handoff

English | [中文](task-handoff.zh.md)

Executor → platform. Frame type: `TASK_HANDOFF`.

Identify the source with `workitemId` and `dispatchId` and propose a recipient with `to` and `toType`. Read the selected recipient from `TASK_HANDOFF_RESULT`.

## Message fields

| Field | JSON type | Description |
| --- | --- | --- |
| `type` | String | Fixed value `TASK_HANDOFF`. |
| `dispatchId` | Integer | ID of this dispatch attempt for the work item. |
| `workitemId` | Integer | ID of the work item. |
| `to` | String | Proposed recipient of the handoff. |
| `toType` | String | Type of the proposed recipient. |

## Example frame

```json
{
  "type": "TASK_HANDOFF",
  "dispatchId": 701,
  "workitemId": 91,
  "to": "reviewer",
  "toType": "AGENT"
}
```

## Example reply

See [TASK_HANDOFF_RESULT](task-handoff-result.md) for the reply fields.

```json
{
  "type": "TASK_HANDOFF_RESULT",
  "dispatchId": 701,
  "workitemId": 91,
  "status": "AGENT_DISPATCHED",
  "downstreamDispatchId": 702,
  "targetType": "AGENT",
  "targetRef": 8,
  "message": "handoff accepted"
}
```

## Related protocol

- [TASK_HANDOFF_RESULT](task-handoff-result.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
