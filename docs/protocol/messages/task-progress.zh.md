# 上报进度

[English](task-progress.md) | 中文

执行器 → 平台。帧类型为 `TASK_PROGRESS`。

用 `dispatchId` 关联派单，`eventType` 和 `message` 描述进度。报告工作流步骤时带 `stepOrder` 或 `stepName`。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `TASK_PROGRESS`。 |
| `dispatchId` | 整数 |  | 本次派单 ID，同一工单重新派发时可能不同。 |
| `eventType` | 字符串 |  | 事件种类，决定如何解读事件内容。 |
| `message` | 字符串 |  | 本次操作的可读说明。 |
| `stepOrder` | 整数 | 工作流步骤进度 | 工作流步骤序号。 |
| `stepName` | 字符串 | 工作流步骤进度 | 工作流步骤名称。 |

## 消息示例

```json
{
  "type": "TASK_PROGRESS",
  "dispatchId": 701,
  "eventType": "package.completed",
  "message": "Package ready"
}
```

1.3.0 未为此消息定义单独的 WebSocket 回执。

## 结构化事件

用 `eventType` 填事件类型，用 `message` 填说明，可直接放在帧顶层或放入 `runtimeEvent` 对象。下例的 `package.completed` 表示任务包已就绪。单独发送自由文本 `resultSummary` 不会生成结构化进度事件。

```json
{
  "type": "TASK_PROGRESS",
  "dispatchId": 701,
  "runtimeEvent": {
    "eventType": "package.completed",
    "eventId": "701-package",
    "seq": 1,
    "message": "Package ready"
  }
}
```

`eventId` 用于识别重复上报，`seq` 是事件序号，`message` 是可读说明。 [DispatchService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DispatchService.java)

## 相关协议

- [接收派单](../tasks/dispatch.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
