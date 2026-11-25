# 应用启动配置

[English](executor-launch-command-v1.md) | 中文

心跳能力标识为 `EXECUTOR_LAUNCH_COMMAND_V1`。

支持携带 `launch` 的重启命令。平台随命令发送 `launchConfigVersion`，配置指定运行时版本时还会带上 `targetRuntimeVersion`。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "EXECUTOR_LAUNCH_COMMAND_V1"
  ]
}
```

## 相关协议

- [EXECUTOR_RESTART](../messages/executor-restart.zh.md)

能力定义与使用见[ExecutorRestartService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorRestartService.java)。
