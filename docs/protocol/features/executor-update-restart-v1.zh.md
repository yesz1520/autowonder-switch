# 更新后重启

[English](executor-update-restart-v1.md) | 中文

心跳能力标识为 `EXECUTOR_UPDATE_RESTART_V1`。

支持 `update: true` 的重启请求。更新和重启进度通过重启结果帧上报。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "EXECUTOR_UPDATE_RESTART_V1"
  ]
}
```

## 相关协议

- [EXECUTOR_RESTART](../messages/executor-restart.zh.md)

能力定义与使用见[ExecutorRestartService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorRestartService.java)。
