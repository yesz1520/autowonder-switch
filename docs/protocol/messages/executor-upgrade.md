# Request an upgrade

English | [中文](executor-upgrade.zh.md)

Platform → executor. Frame type: `EXECUTOR_UPGRADE`.

Read the upgrade request for `executorId` and `requestId`. The target is `targetVersion`. The frame also carries `issuedAt` and `resultReceiptAck`.

## Message fields

| Field | JSON type | Description |
| --- | --- | --- |
| `type` | String | Fixed value `EXECUTOR_UPGRADE`. |
| `executorId` | Integer | ID of the executor on this connection. |
| `requestId` | String | String identifier correlating the request with subsequent messages. |
| `targetVersion` | String | Target version for the upgrade. |
| `issuedAt` | String | Time the platform issued the command. |
| `resultReceiptAck` | Boolean | Whether the platform supports receipts for terminal upgrade results. |

## Example frame

```json
{
  "type": "EXECUTOR_UPGRADE",
  "executorId": 42,
  "requestId": "request-01",
  "targetVersion": "1.3.0",
  "issuedAt": "2026-11-25T08:00:00Z",
  "resultReceiptAck": true
}
```

## Example reply

See [EXECUTOR_UPGRADE_RESULT](executor-upgrade-result.md) for the reply fields.

```json
{
  "type": "EXECUTOR_UPGRADE_RESULT",
  "requestId": "request-01",
  "phase": "success",
  "receiptId": "receipt-01",
  "currentVersion": "1.3.0"
}
```

## Related protocol

- [EXECUTOR_UPGRADE_RESULT](executor-upgrade-result.md)
- [EXECUTOR_UPDATE_V1](../features/executor-update-v1.md)

Message handling in [ExecutorUpdateService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorUpdateService.java).
