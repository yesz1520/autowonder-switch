# 远程升级

[English](executor-update-v1.md) | 中文

心跳能力标识为 `EXECUTOR_UPDATE_V1`。

支持平台指定目标版本的远程升级。执行器回报升级阶段及实际启动的版本。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "EXECUTOR_UPDATE_V1"
  ]
}
```

## 相关协议

- [EXECUTOR_UPGRADE](../messages/executor-upgrade.zh.md)
- [EXECUTOR_UPGRADE_RESULT](../messages/executor-upgrade-result.zh.md)

能力定义与使用见[ExecutorUpdateService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorUpdateService.java)。
