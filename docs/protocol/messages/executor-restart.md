# Request a restart

English | [中文](executor-restart.zh.md)

Platform → executor. Frame type: `EXECUTOR_RESTART`.

Identify the request by `executorId` and `requestId`. Handle the restart according to `update`, optional `restartMode` and `launch`. The issue time is `issuedAt`.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `EXECUTOR_RESTART`. |
| `executorId` | Integer |  | ID of the executor on this connection. |
| `requestId` | String |  | String identifier correlating the request with subsequent messages. |
| `issuedAt` | String |  | Time the platform issued the command. |
| `update` | Boolean |  | Whether to update before restarting. |
| `status` | String |  | State when the platform issues the restart. |
| `restartMode` | String | When configured for restart | WHEN_IDLE waits for idle. FORCE requests a forced restart. |
| `launch` | Object | When configured for restart | Launch configuration sent by the platform. |
| `launchConfigVersion` | Integer | When configured for restart | Version of the launch configuration. |
| `targetRuntimeVersion` | String | When configured for restart | Runtime version specified in the launch configuration. |
| `previousStartedAt` | Integer |  | Previously recorded startup time in Unix milliseconds, if known. |
| `requestedBy` | Integer |  | User ID that requested the restart. |
| `messageOrigin` | String |  | Origin of the restart status message. |
| `message` | String |  | Human-readable description of the operation. |

## Example frame

```json
{
  "type": "EXECUTOR_RESTART",
  "executorId": 42,
  "requestId": "request-01",
  "issuedAt": "2026-11-25T08:00:00Z",
  "update": false,
  "status": "REQUESTED"
}
```

## Example reply

See [EXECUTOR_RESTART_RESULT](executor-restart-result.md) for the reply fields.

```json
{
  "type": "EXECUTOR_RESTART_RESULT",
  "requestId": "request-01",
  "status": "RESTARTING",
  "message": "Restarting"
}
```

## Related protocol

- [EXECUTOR_RESTART_RESULT](executor-restart-result.md)
- [Heartbeat](../heartbeat.md)
- [EXECUTOR_RESTART_V1](../features/executor-restart-v1.md)
- [EXECUTOR_UPDATE_RESTART_V1](../features/executor-update-restart-v1.md)
- [EXECUTOR_LAUNCH_COMMAND_V1](../features/executor-launch-command-v1.md)
- [EXECUTOR_RESTART_POLICY_V1](../features/executor-restart-policy-v1.md)

Message handling in [ExecutorRestartService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorRestartService.java).
