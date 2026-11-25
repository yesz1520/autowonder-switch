# 上报重启进度

[English](executor-restart-result.md) | 中文

执行器 → 平台。帧类型为 `EXECUTOR_RESTART_RESULT`。

回传 `requestId` 和 `status`，可附 `message`。状态为 `WAITING_IDLE`、`UPDATING`、`RESTARTING` 或 `FAILED`。启动完成由新进程的心跳确认。

## 消息参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `type` | 字符串 | 固定为 `EXECUTOR_RESTART_RESULT`。 |
| `requestId` | 字符串 | 关联请求与后续消息的字符串 ID。 |
| `status` | 字符串 | 重启阶段：WAITING_IDLE、UPDATING、RESTARTING 或 FAILED。 |
| `message` | 字符串 | 本次操作的可读说明。 |

## 消息示例

```json
{
  "type": "EXECUTOR_RESTART_RESULT",
  "requestId": "request-01",
  "status": "RESTARTING",
  "message": "Restarting"
}
```

## 相关协议

- [EXECUTOR_RESTART](executor-restart.zh.md)
- [心跳](../heartbeat.zh.md)
- [EXECUTOR_RESTART_V1](../features/executor-restart-v1.zh.md)

消息处理见[ExecutorRestartService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorRestartService.java)。
