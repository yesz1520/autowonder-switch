# 返回交接决定

[English](task-handoff-result.md) | 中文

平台 → 执行器。帧类型为 `TASK_HANDOFF_RESULT`。

用 `workitemId`、`dispatchId` 匹配交接请求。读取 `status`，并按结果读取 `downstreamDispatchId`、`targetType`、`targetRef`、`reasonCode` 或 `message`。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `TASK_HANDOFF_RESULT`。 |
| `dispatchId` | 整数 |  | 本次派单 ID，同一工单重新派发时可能不同。 |
| `workitemId` | 整数 |  | 派单所属工单的 ID。 |
| `status` | 字符串 |  | 交接决定，如 AGENT_DISPATCHED、HUMAN_ASSIGNED 或 REJECTED。 |
| `downstreamDispatchId` | 整数 |  | 交接成功后新建的下游派单 ID。 |
| `targetType` | 字符串 |  | 实际交接目标的类型。 |
| `targetRef` | 整数 |  | 交接决定选中的目标 ID。 |
| `message` | 字符串 |  | 本次操作的可读说明。 |
| `reasonCode` | 字符串 | 有原因代码时 | 服务端给出的原因代码。 |

## 消息示例

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

- [TASK_HANDOFF](task-handoff.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
