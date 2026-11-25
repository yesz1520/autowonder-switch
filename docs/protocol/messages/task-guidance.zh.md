# 下发任务指导

[English](task-guidance.md) | 中文

平台 → 执行器。帧类型为 `TASK_GUIDANCE`。

平台用 `guidanceId` 标识指导内容。帧含 `executorId`、`dispatchId`、`workitemId`、`targetAgentId`、`content` 和 `mode: "SAFE_STEER"`。

## 消息参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `type` | 字符串 | 固定为 `TASK_GUIDANCE`。 |
| `executorId` | 整数 | 该连接对应的执行器 ID。 |
| `guidanceId` | 整数 | 指导或追加指令的 ID，回复时原样带回。 |
| `dispatchId` | 整数 | 本次派单 ID，同一工单重新派发时可能不同。 |
| `workitemId` | 整数 | 派单所属工单的 ID。 |
| `targetAgentId` | 整数 | 指导所针对的智能体 ID。 |
| `content` | 字符串 | 向执行器投递的内容。 |
| `mode` | 字符串 | 本次任务指导的投递模式。 |

## 消息示例

```json
{
  "type": "TASK_GUIDANCE",
  "executorId": 42,
  "guidanceId": 501,
  "dispatchId": 701,
  "workitemId": 91,
  "targetAgentId": 7,
  "content": "Review the changes",
  "mode": "SAFE_STEER"
}
```

## 对应回包

回包参数见 [TASK_GUIDANCE_ACK](task-guidance-ack.zh.md)。

```json
{
  "type": "TASK_GUIDANCE_ACK",
  "guidanceId": 501,
  "dispatchId": 701,
  "status": "APPLIED",
  "replyMarkdown": "Review complete"
}
```

## 相关协议

- [TASK_GUIDANCE_ACK](task-guidance-ack.zh.md)

消息处理见[WsGuidanceTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsGuidanceTransport.java)。
