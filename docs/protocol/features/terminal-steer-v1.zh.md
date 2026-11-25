# 运行中追加指令

[English](terminal-steer-v1.md) | 中文

心跳能力标识为 `terminal_steer_v1`。

运行中的派单可以接收追加指令，执行器按 `guidanceId` 回复应用结果。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "terminal_steer_v1"
  ]
}
```

## 相关协议

- [TASK_STEER](../messages/task-steer.zh.md)
- [TASK_STEER_ACK](../messages/task-steer-ack.zh.md)

能力定义与使用见[DispatchGuidanceService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DispatchGuidanceService.java)。
