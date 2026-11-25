# 请求交接

[English](task-handoff.md) | 中文

执行器 → 平台。帧类型为 `TASK_HANDOFF`。

用 `workitemId` 和 `dispatchId` 指明交接来源，`to` 和 `toType` 指明建议的接收方。实际接收方以 `TASK_HANDOFF_RESULT` 为准。

## 消息参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `type` | 字符串 | 固定为 `TASK_HANDOFF`。 |
| `dispatchId` | 整数 | 本次派单 ID，同一工单重新派发时可能不同。 |
| `workitemId` | 整数 | 派单所属工单的 ID。 |
| `to` | 字符串 | 建议接收交接的目标。 |
| `toType` | 字符串 | 交接目标的类型。 |

## 消息示例

```json
{
  "type": "TASK_HANDOFF",
  "dispatchId": 701,
  "workitemId": 91,
  "to": "reviewer",
  "toType": "AGENT"
}
```

## 对应回包

回包参数见 [TASK_HANDOFF_RESULT](task-handoff-result.zh.md)。

```json
{
  "type": "TASK_HANDOFF_RESULT",
  "dispatchId": 701,
  "workitemId": 91,
  "status": "AGENT_DISPATCHED",
  "downstreamDispatchId": 702,
  "targetType": "AGENT",
  "targetRef": 8,
  "message": "handoff accepted"
}
```

## 相关协议

- [TASK_HANDOFF_RESULT](task-handoff-result.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
