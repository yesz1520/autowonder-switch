# 确认收到升级结果

[English](executor-upgrade-result-ack.md) | 中文

平台 → 执行器。帧类型为 `EXECUTOR_UPGRADE_RESULT_ACK`。

平台收到带回执 ID 的最终结果后，带回 `requestId` 和 `receiptId`。执行器据此确认对应的结果投递。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `EXECUTOR_UPGRADE_RESULT_ACK`。 |
| `requestId` | 字符串 |  | 关联请求与后续消息的字符串 ID。 |
| `receiptId` | 字符串 | 探测或确认升级结果时 | 升级结果的回执标识，探测与最终投递均按此标识匹配。 |

## 消息示例

```json
{
  "type": "EXECUTOR_UPGRADE_RESULT_ACK",
  "requestId": "request-01",
  "receiptId": "receipt-01"
}
```

## 相关协议

- [EXECUTOR_UPGRADE_RESULT](executor-upgrade-result.zh.md)

消息处理见[ExecutorUpdateService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorUpdateService.java)。
