# Steer a running task

English | [中文](task-steer.zh.md)

Platform → executor. Frame type: `TASK_STEER`.

Deliver additional `content` to a running dispatch. Use `executorId` and `dispatchId` to find the task and `guidanceId` to identify the instruction.

## Message fields

| Field | JSON type | Description |
| --- | --- | --- |
| `type` | String | Fixed value `TASK_STEER`. |
| `executorId` | Integer | ID of the executor on this connection. |
| `guidanceId` | Integer | Guidance or steering instruction ID, echoed in the reply. |
| `dispatchId` | Integer | ID of this dispatch attempt for the work item. |
| `content` | String | Content delivered to the executor. |

## Example frame

```json
{
  "type": "TASK_STEER",
  "executorId": 42,
  "guidanceId": 501,
  "dispatchId": 701,
  "content": "Review the changes"
}
```

## Example reply

See [TASK_STEER_ACK](task-steer-ack.md) for the reply fields.

```json
{
  "type": "TASK_STEER_ACK",
  "dispatchId": 701,
  "guidanceId": 501,
  "status": "APPLIED"
}
```

## Related protocol

- [TASK_STEER_ACK](task-steer-ack.md)
- [terminal_steer_v1](../features/terminal-steer-v1.md)

Message handling in [DispatchGuidanceService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DispatchGuidanceService.java).
