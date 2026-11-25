# Acknowledge steering

English | [中文](task-steer-ack.zh.md)

Executor → platform. Frame type: `TASK_STEER_ACK`.

Echo `dispatchId` and `guidanceId`. Send `status: "APPLIED"` after applying the instruction, or `FAILED` with an `error`.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `TASK_STEER_ACK`. |
| `dispatchId` | Integer |  | ID of this dispatch attempt for the work item. |
| `guidanceId` | Integer |  | Guidance or steering instruction ID, echoed in the reply. |
| `status` | String |  | Steering outcome, APPLIED or FAILED. |
| `error` | String | On failure | Description of the failure. |

## Example frame

```json
{
  "type": "TASK_STEER_ACK",
  "dispatchId": 701,
  "guidanceId": 501,
  "status": "APPLIED"
}
```

## Related protocol

- [TASK_STEER](task-steer.md)
- [terminal_steer_v1](../features/terminal-steer-v1.md)

Message handling in [DispatchGuidanceService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DispatchGuidanceService.java).
