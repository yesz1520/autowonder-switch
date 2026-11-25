# Send task guidance

English | [中文](task-guidance.zh.md)

Platform → executor. Frame type: `TASK_GUIDANCE`.

The platform identifies guidance by `guidanceId`. The frame includes `executorId`, `dispatchId`, `workitemId`, `targetAgentId`, `content` and `mode: "SAFE_STEER"`.

## Message fields

| Field | JSON type | Description |
| --- | --- | --- |
| `type` | String | Fixed value `TASK_GUIDANCE`. |
| `executorId` | Integer | ID of the executor on this connection. |
| `guidanceId` | Integer | Guidance or steering instruction ID, echoed in the reply. |
| `dispatchId` | Integer | ID of this dispatch attempt for the work item. |
| `workitemId` | Integer | ID of the work item. |
| `targetAgentId` | Integer | Agent targeted by the guidance. |
| `content` | String | Content delivered to the executor. |
| `mode` | String | Delivery mode for task guidance. |

## Example frame

```json
{
  "type": "TASK_GUIDANCE",
  "executorId": 42,
  "guidanceId": 501,
  "dispatchId": 701,
  "workitemId": 91,
  "targetAgentId": 7,
  "content": "Review the changes",
  "mode": "SAFE_STEER"
}
```

## Example reply

See [TASK_GUIDANCE_ACK](task-guidance-ack.md) for the reply fields.

```json
{
  "type": "TASK_GUIDANCE_ACK",
  "guidanceId": 501,
  "dispatchId": 701,
  "status": "APPLIED",
  "replyMarkdown": "Review complete"
}
```

## Related protocol

- [TASK_GUIDANCE_ACK](task-guidance-ack.md)

Message handling in [WsGuidanceTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsGuidanceTransport.java).
