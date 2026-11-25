# 轻量检查点恢复

[English](checkpoint-lightweight-v1.md) | 中文

心跳能力标识为 `CHECKPOINT_LIGHTWEIGHT_V1`。

声明可在恢复派单时读取轻量检查点。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "CHECKPOINT_LIGHTWEIGHT_V1"
  ]
}
```

## 相关协议

- [暂停与恢复](../tasks/recovery.zh.md)

能力定义与使用见[DispatchService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DispatchService.java)。
