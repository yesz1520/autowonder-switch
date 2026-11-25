# 报告忙碌

[English](task-busy.md) | 中文

执行器 → 平台。帧类型为 `TASK_BUSY`。

无法接单时带回 `dispatchId`，可用 `reason` 说明原因。

## 消息参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `type` | 字符串 | 固定为 `TASK_BUSY`。 |
| `dispatchId` | 整数 | 本次派单 ID，同一工单重新派发时可能不同。 |
| `reason` | 字符串 | 执行器暂时无法接单的原因。 |

## 消息示例

```json
{
  "type": "TASK_BUSY",
  "dispatchId": 701,
  "reason": "capacity reached"
}
```

1.3.0 未为此消息定义单独的 WebSocket 回执。

## 相关协议

- [TASK_DISPATCH](task-dispatch.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
