# Confirm upgrade receipt support

English | [中文](executor-upgrade-result-support.zh.md)

Platform → executor. Frame type: `EXECUTOR_UPGRADE_RESULT_SUPPORT`.

In response to `receipt_probe`, the platform echoes `requestId` and `receiptId` to confirm support on this connection. The executor still needs to send its final result.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `EXECUTOR_UPGRADE_RESULT_SUPPORT`. |
| `requestId` | String |  | String identifier correlating the request with subsequent messages. |
| `receiptId` | String | Receipt probe or result delivery | Identifier used to correlate upgrade result delivery or a receipt probe. |

## Example frame

```json
{
  "type": "EXECUTOR_UPGRADE_RESULT_SUPPORT",
  "requestId": "request-01",
  "receiptId": "receipt-01"
}
```

## Related protocol

- [EXECUTOR_UPGRADE_RESULT](executor-upgrade-result.md)

Message handling in [ExecutorUpdateService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorUpdateService.java).
