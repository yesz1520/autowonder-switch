# 重启时机

[English](executor-restart-policy-v1.md) | 中文

心跳能力标识为 `EXECUTOR_RESTART_POLICY_V1`。

支持 `restartMode`。取值 `WHEN_IDLE` 表示等待空闲，`FORCE` 表示强制重启。此能力与对应的重启能力一起声明。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "EXECUTOR_RESTART_V1",
    "EXECUTOR_RESTART_POLICY_V1"
  ]
}
```

## 相关协议

- [EXECUTOR_RESTART](../messages/executor-restart.zh.md)

能力定义与使用见[ExecutorRestartService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorRestartService.java)。
