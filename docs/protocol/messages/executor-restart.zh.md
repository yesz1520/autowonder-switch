# 请求重启

[English](executor-restart.md) | 中文

平台 → 执行器。帧类型为 `EXECUTOR_RESTART`。

用 `executorId` 和 `requestId` 标识请求。按 `update`、可选的 `restartMode` 和 `launch` 处理本次重启，`issuedAt` 是签发时间。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `EXECUTOR_RESTART`。 |
| `executorId` | 整数 |  | 该连接对应的执行器 ID。 |
| `requestId` | 字符串 |  | 关联请求与后续消息的字符串 ID。 |
| `issuedAt` | 字符串 |  | 平台签发此次命令的时间。 |
| `update` | 布尔值 |  | 重启前是否执行更新。 |
| `status` | 字符串 |  | 平台签发重启时的状态。 |
| `restartMode` | 字符串 | 指定对应重启配置时 | WHEN_IDLE 表示空闲时重启，FORCE 表示强制重启。 |
| `launch` | 对象 | 指定对应重启配置时 | 平台下发的启动配置。 |
| `launchConfigVersion` | 整数 | 指定对应重启配置时 | 本次下发的启动配置版本。 |
| `targetRuntimeVersion` | 字符串 | 指定对应重启配置时 | 启动配置要求运行的版本。 |
| `previousStartedAt` | 整数 |  | 重启前记录的启动时间，Unix 毫秒，可为空。 |
| `requestedBy` | 整数 |  | 发起重启的用户 ID。 |
| `messageOrigin` | 字符串 |  | 重启状态消息的来源。 |
| `message` | 字符串 |  | 本次操作的可读说明。 |

## 消息示例

```json
{
  "type": "EXECUTOR_RESTART",
  "executorId": 42,
  "requestId": "request-01",
  "issuedAt": "2026-11-25T08:00:00Z",
  "update": false,
  "status": "REQUESTED"
}
```

## 对应回包

回包参数见 [EXECUTOR_RESTART_RESULT](executor-restart-result.zh.md)。

```json
{
  "type": "EXECUTOR_RESTART_RESULT",
  "requestId": "request-01",
  "status": "RESTARTING",
  "message": "Restarting"
}
```

## 相关协议

- [EXECUTOR_RESTART_RESULT](executor-restart-result.zh.md)
- [心跳](../heartbeat.zh.md)
- [EXECUTOR_RESTART_V1](../features/executor-restart-v1.zh.md)
- [EXECUTOR_UPDATE_RESTART_V1](../features/executor-update-restart-v1.zh.md)
- [EXECUTOR_LAUNCH_COMMAND_V1](../features/executor-launch-command-v1.zh.md)
- [EXECUTOR_RESTART_POLICY_V1](../features/executor-restart-policy-v1.zh.md)

消息处理见[ExecutorRestartService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorRestartService.java)。
