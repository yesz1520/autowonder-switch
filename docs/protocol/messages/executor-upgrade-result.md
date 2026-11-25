# Report upgrade progress or outcome

English | [中文](executor-upgrade-result.zh.md)

Executor → platform. Frame type: `EXECUTOR_UPGRADE_RESULT`.

Echo `requestId`. Progress phases are `accepted`, `draining`, `downloading` and `applying`. Report `success` with `currentVersion` or `failed` with `error`. Use `phase: "receipt_probe"` with `receiptId` to probe receipt support. A final failure carrying a receipt ID also requires `terminal: true`.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `EXECUTOR_UPGRADE_RESULT`. |
| `requestId` | String |  | String identifier correlating the request with subsequent messages. |
| `phase` | String |  | Upgrade phase, or receipt_probe to negotiate receipts. |
| `receiptId` | String | Receipt probe or result delivery | Identifier used to correlate upgrade result delivery or a receipt probe. |
| `currentVersion` | String | On upgrade success | Version actually running after startup. |
| `error` | String | On failure | Description of the failure. |
| `terminal` | Boolean | On terminal failure | Whether this is a terminal failure reported with a receipt. |

## Example frame

```json
{
  "type": "EXECUTOR_UPGRADE_RESULT",
  "requestId": "request-01",
  "phase": "success",
  "receiptId": "receipt-01",
  "currentVersion": "1.3.0"
}
```

## Example reply

See [EXECUTOR_UPGRADE_RESULT_ACK](executor-upgrade-result-ack.md) for the reply fields.

```json
{
  "type": "EXECUTOR_UPGRADE_RESULT_ACK",
  "requestId": "request-01",
  "receiptId": "receipt-01"
}
```

## Related protocol

- [EXECUTOR_UPGRADE_RESULT_SUPPORT](executor-upgrade-result-support.md)
- [EXECUTOR_UPGRADE_RESULT_ACK](executor-upgrade-result-ack.md)
- [EXECUTOR_UPDATE_V1](../features/executor-update-v1.md)

Message handling in [ExecutorUpdateService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorUpdateService.java).
