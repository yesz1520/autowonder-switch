# 上报升级进度或结果

[English](executor-upgrade-result.md) | 中文

执行器 → 平台。帧类型为 `EXECUTOR_UPGRADE_RESULT`。

回传 `requestId`。进度 `phase` 为 `accepted`、`draining`、`downloading` 或 `applying`。成功填 `success` 和 `currentVersion`，失败填 `failed` 和 `error`。探测回执支持时填 `receipt_probe` 和 `receiptId`。携带回执 ID 的最终失败还须带 `terminal: true`。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `EXECUTOR_UPGRADE_RESULT`。 |
| `requestId` | 字符串 |  | 关联请求与后续消息的字符串 ID。 |
| `phase` | 字符串 |  | 升级阶段，或用于探测回执支持的 receipt_probe。 |
| `receiptId` | 字符串 | 探测或确认升级结果时 | 升级结果的回执标识，探测与最终投递均按此标识匹配。 |
| `currentVersion` | 字符串 | 升级成功时 | 启动后实际运行的版本。 |
| `error` | 字符串 | 失败时 | 操作失败的原因。 |
| `terminal` | 布尔值 | 上报最终失败时 | 失败是否属于带回执的最终失败。 |

## 消息示例

```json
{
  "type": "EXECUTOR_UPGRADE_RESULT",
  "requestId": "request-01",
  "phase": "success",
  "receiptId": "receipt-01",
  "currentVersion": "1.3.0"
}
```

## 对应回包

回包参数见 [EXECUTOR_UPGRADE_RESULT_ACK](executor-upgrade-result-ack.zh.md)。

```json
{
  "type": "EXECUTOR_UPGRADE_RESULT_ACK",
  "requestId": "request-01",
  "receiptId": "receipt-01"
}
```

## 相关协议

- [EXECUTOR_UPGRADE_RESULT_SUPPORT](executor-upgrade-result-support.zh.md)
- [EXECUTOR_UPGRADE_RESULT_ACK](executor-upgrade-result-ack.zh.md)
- [EXECUTOR_UPDATE_V1](../features/executor-update-v1.zh.md)

消息处理见[ExecutorUpdateService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorUpdateService.java)。
