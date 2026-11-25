# 确认追加指令

[English](task-steer-ack.md) | 中文

执行器 → 平台。帧类型为 `TASK_STEER_ACK`。

回传 `dispatchId` 和 `guidanceId`。指令已应用时填 `status: "APPLIED"`，失败时填 `FAILED` 并说明 `error`。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `TASK_STEER_ACK`。 |
| `dispatchId` | 整数 |  | 本次派单 ID，同一工单重新派发时可能不同。 |
| `guidanceId` | 整数 |  | 指导或追加指令的 ID，回复时原样带回。 |
| `status` | 字符串 |  | 追加指令是否已应用，取 APPLIED 或 FAILED。 |
| `error` | 字符串 | 失败时 | 操作失败的原因。 |

## 消息示例

```json
{
  "type": "TASK_STEER_ACK",
  "dispatchId": 701,
  "guidanceId": 501,
  "status": "APPLIED"
}
```

## 相关协议

- [TASK_STEER](task-steer.zh.md)
- [terminal_steer_v1](../features/terminal-steer-v1.zh.md)

消息处理见[DispatchGuidanceService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DispatchGuidanceService.java)。
