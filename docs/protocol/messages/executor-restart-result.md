# Report restart progress

English | [中文](executor-restart-result.zh.md)

Executor → platform. Frame type: `EXECUTOR_RESTART_RESULT`.

Echo `requestId` and `status` with an optional `message`. Status is `WAITING_IDLE`, `UPDATING`, `RESTARTING` or `FAILED`. The new process confirms startup in its heartbeat.

## Message fields

| Field | JSON type | Description |
| --- | --- | --- |
| `type` | String | Fixed value `EXECUTOR_RESTART_RESULT`. |
| `requestId` | String | String identifier correlating the request with subsequent messages. |
| `status` | String | Restart stage: WAITING_IDLE, UPDATING, RESTARTING or FAILED. |
| `message` | String | Human-readable description of the operation. |

## Example frame

```json
{
  "type": "EXECUTOR_RESTART_RESULT",
  "requestId": "request-01",
  "status": "RESTARTING",
  "message": "Restarting"
}
```

## Related protocol

- [EXECUTOR_RESTART](executor-restart.md)
- [Heartbeat](../heartbeat.md)
- [EXECUTOR_RESTART_V1](../features/executor-restart-v1.md)

Message handling in [ExecutorRestartService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorRestartService.java).
