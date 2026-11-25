# 执行清单

[English](execution-inventory.md) | 中文

执行器在[心跳](heartbeat.zh.md)中上报当前容量、派单和会话轮次。清单字段与 `type` 同级。

## 清单字段

声明 `dispatch_inventory_v1` 后，每次心跳都要发送下表的五个必填字段，即使执行器当前空闲或清单尚未就绪。ID 使用不带引号的正整数，同一数组中不能重复。必填字段缺失、类型错误或清单不符合约束时，平台记录协议错误并关闭连接。

| 字段 | JSON 类型 | 何时发送 | 描述 |
|---|---|---|---|
| `maxConcurrentDispatches` | 整数 | 必填 | 执行器的总并发容量，取值 1–50。即使握手时已声明，也要随清单发送。 |
| `dispatchInventoryReady` | 布尔值 | 必填 | 清单完整、可供平台核对时填 `true`。执行器仍在恢复本地状态或无法列出全部持有的派单时填 `false`。 |
| `runningDispatchIds` | 整数数组 | 必填 | 正在执行的派单 ID，最多 50 个。没有正在执行的派单时填 `[]`。 |
| `ownedDispatchIds` | 整数数组 | 必填 | 执行器仍持有的派单 ID，最多 1,000 个。必须包含全部 `runningDispatchIds`，也包括已结束执行、仍在等待结果回执的派单。没有时填 `[]`。 |
| `runningConversationTurnIds` | 整数数组 | 必填 | 执行器仍持有的会话轮次 ID。填写轮次的 `turnId`，最多 50 个。没有时填 `[]`。 |
| `dispatchInventoryError` | 字符串 | 持有的派单超过 1,000 个时 | 填 `OWNED_DISPATCH_LIMIT_EXCEEDED`，将 `dispatchInventoryReady` 设为 `false`，并在两份派单数组中都只填正在执行的派单 ID。正常情况下省略。 |
| `computingConversationTurnIds` | 整数数组 | 声明 `CONVERSATION_COMPUTE_INVENTORY_V1` 时 | 填写持有的轮次中正在计算的轮次 ID，最多 50 个。传 `[]` 表示当前没有轮次在计算。省略、格式错误或包含未持有的轮次时，平台将计算状态视为未知，其余有效的心跳字段仍可使用。 |

未声明 `dispatch_inventory_v1` 时，平台沿用握手时的容量。执行器仍可发送 `runningDispatchIds` 和 `runningConversationTurnIds`。旧客户端也可用 `runningConversationTurns` 对象数组上报轮次，每项包含 `turnId`。省略两种轮次字段表示未报告轮次状态，显式传空数组表示当前没有持有的轮次。

## 心跳示例

派单 701 已结束执行，仍在等待结果回执，因此只出现在 `ownedDispatchIds` 中。

```json
{
  "type": "HEARTBEAT",
  "startedAt": "2026-11-25T08:00:00Z",
  "version": "1.2.3",
  "model": "example-model",
  "protocolFeatures": [
    "dispatch_inventory_v1",
    "CONVERSATION_COMPUTE_INVENTORY_V1"
  ],
  "maxConcurrentDispatches": 2,
  "dispatchInventoryReady": true,
  "runningDispatchIds": [],
  "ownedDispatchIds": [
    701
  ],
  "runningConversationTurnIds": [],
  "computingConversationTurnIds": []
}
```

## 相关源码

- 解析心跳字段并校验清单。 [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)
