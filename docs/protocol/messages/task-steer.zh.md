# 向运行中的任务追加指令

[English](task-steer.md) | 中文

平台 → 执行器。帧类型为 `TASK_STEER`。

向正在运行的派单追加 `content`。按 `executorId`、`dispatchId` 定位任务，按 `guidanceId` 识别这条指令。

## 消息参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `type` | 字符串 | 固定为 `TASK_STEER`。 |
| `executorId` | 整数 | 该连接对应的执行器 ID。 |
| `guidanceId` | 整数 | 指导或追加指令的 ID，回复时原样带回。 |
| `dispatchId` | 整数 | 本次派单 ID，同一工单重新派发时可能不同。 |
| `content` | 字符串 | 向执行器投递的内容。 |

## 消息示例

```json
{
  "type": "TASK_STEER",
  "executorId": 42,
  "guidanceId": 501,
  "dispatchId": 701,
  "content": "Review the changes"
}
```

## 对应回包

回包参数见 [TASK_STEER_ACK](task-steer-ack.zh.md)。

```json
{
  "type": "TASK_STEER_ACK",
  "dispatchId": 701,
  "guidanceId": 501,
  "status": "APPLIED"
}
```

## 相关协议

- [TASK_STEER_ACK](task-steer-ack.zh.md)
- [terminal_steer_v1](../features/terminal-steer-v1.zh.md)

消息处理见[DispatchGuidanceService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DispatchGuidanceService.java)。
