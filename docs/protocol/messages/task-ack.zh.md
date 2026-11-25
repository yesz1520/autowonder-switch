# 确认接单

[English](task-ack.md) | 中文

执行器 → 平台。帧类型为 `TASK_ACK`。

接单后回传原派单的 `dispatchId`。任务进度和最终结果随后单独上报。

## 消息参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `type` | 字符串 | 固定为 `TASK_ACK`。 |
| `dispatchId` | 整数 | 本次派单 ID，同一工单重新派发时可能不同。 |

## 消息示例

```json
{
  "type": "TASK_ACK",
  "dispatchId": 701
}
```

1.3.0 未为此消息定义单独的 WebSocket 回执。

## 相关协议

- [TASK_DISPATCH](task-dispatch.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
