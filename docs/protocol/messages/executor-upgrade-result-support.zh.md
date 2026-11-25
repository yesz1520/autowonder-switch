# 确认支持升级结果回执

[English](executor-upgrade-result-support.md) | 中文

平台 → 执行器。帧类型为 `EXECUTOR_UPGRADE_RESULT_SUPPORT`。

收到 `receipt_probe` 后，平台带回 `requestId` 和 `receiptId`，表明当前连接支持升级结果回执。执行器仍需发送最终结果。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `EXECUTOR_UPGRADE_RESULT_SUPPORT`。 |
| `requestId` | 字符串 |  | 关联请求与后续消息的字符串 ID。 |
| `receiptId` | 字符串 | 探测或确认升级结果时 | 升级结果的回执标识，探测与最终投递均按此标识匹配。 |

## 消息示例

```json
{
  "type": "EXECUTOR_UPGRADE_RESULT_SUPPORT",
  "requestId": "request-01",
  "receiptId": "receipt-01"
}
```

## 相关协议

- [EXECUTOR_UPGRADE_RESULT](executor-upgrade-result.zh.md)

消息处理见[ExecutorUpdateService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorUpdateService.java)。
