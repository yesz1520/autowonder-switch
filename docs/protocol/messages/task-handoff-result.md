# Return a handoff decision

English | [中文](task-handoff-result.zh.md)

Platform → executor. Frame type: `TASK_HANDOFF_RESULT`.

Match the request by `workitemId` and `dispatchId`. Read `status` and, as applicable, `downstreamDispatchId`, `targetType`, `targetRef`, `reasonCode` or `message`.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `TASK_HANDOFF_RESULT`. |
| `dispatchId` | Integer |  | ID of this dispatch attempt for the work item. |
| `workitemId` | Integer |  | ID of the work item. |
| `status` | String |  | Decision: AGENT_DISPATCHED, HUMAN_ASSIGNED or REJECTED. |
| `downstreamDispatchId` | Integer |  | ID of the downstream dispatch created by the handoff. |
| `targetType` | String |  | Type of the selected recipient. |
| `targetRef` | Integer |  | Recipient ID chosen for the handoff. |
| `message` | String |  | Human-readable description of the operation. |
| `reasonCode` | String | When a reason code is available | Machine-readable reason for the outcome. |

## Example frame

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

- [TASK_HANDOFF](task-handoff.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
