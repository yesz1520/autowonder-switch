# Acknowledge an upgrade outcome

English | [中文](executor-upgrade-result-ack.zh.md)

Platform → executor. Frame type: `EXECUTOR_UPGRADE_RESULT_ACK`.

The platform echoes `requestId` and `receiptId` after receiving a final result with a receipt ID. Match both values to confirm delivery.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `EXECUTOR_UPGRADE_RESULT_ACK`. |
| `requestId` | String |  | String identifier correlating the request with subsequent messages. |
| `receiptId` | String | Receipt probe or result delivery | Identifier used to correlate upgrade result delivery or a receipt probe. |

## Example frame

```json
{
  "type": "EXECUTOR_UPGRADE_RESULT_ACK",
  "requestId": "request-01",
  "receiptId": "receipt-01"
}
```

## Related protocol

- [EXECUTOR_UPGRADE_RESULT](executor-upgrade-result.md)

Message handling in [ExecutorUpdateService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorUpdateService.java).
