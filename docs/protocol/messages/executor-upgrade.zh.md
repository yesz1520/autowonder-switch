# 请求升级

[English](executor-upgrade.md) | 中文

平台 → 执行器。帧类型为 `EXECUTOR_UPGRADE`。

按 `executorId` 和 `requestId` 接收升级请求，`targetVersion` 指定目标版本。帧还带 `issuedAt` 和 `resultReceiptAck`。

## 消息参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `type` | 字符串 | 固定为 `EXECUTOR_UPGRADE`。 |
| `executorId` | 整数 | 该连接对应的执行器 ID。 |
| `requestId` | 字符串 | 关联请求与后续消息的字符串 ID。 |
| `targetVersion` | 字符串 | 本次远程升级的目标版本。 |
| `issuedAt` | 字符串 | 平台签发此次命令的时间。 |
| `resultReceiptAck` | 布尔值 | 平台是否支持升级最终结果回执。 |

## 消息示例

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

## 对应回包

回包参数见 [EXECUTOR_UPGRADE_RESULT](executor-upgrade-result.zh.md)。

```json
{
  "type": "EXECUTOR_UPGRADE_RESULT",
  "requestId": "request-01",
  "phase": "success",
  "receiptId": "receipt-01",
  "currentVersion": "1.3.0"
}
```

## 相关协议

- [EXECUTOR_UPGRADE_RESULT](executor-upgrade-result.zh.md)
- [EXECUTOR_UPDATE_V1](../features/executor-update-v1.zh.md)

消息处理见[ExecutorUpdateService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorUpdateService.java)。
