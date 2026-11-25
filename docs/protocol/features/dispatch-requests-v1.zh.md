# 派单中的人工请求

[English](dispatch-requests-v1.md) | 中文

心跳能力标识为 `DISPATCH_REQUESTS_V1`。

声明可处理需要人工响应请求的派单。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "DISPATCH_REQUESTS_V1"
  ]
}
```

## 相关协议

- [TASK_DISPATCH](../messages/task-dispatch.zh.md)

能力定义与使用见[DispatchService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DispatchService.java)。
