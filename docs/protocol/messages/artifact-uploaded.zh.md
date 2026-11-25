# 登记已上传产物

[English](artifact-uploaded.md) | 中文

执行器 → 平台。帧类型为 `ARTIFACT_UPLOADED`。

上报已上传的产物。用 `dispatchId` 关联派单，`name`、`artifactType`、`ossRef`、`size` 和 `metaJson` 描述产物记录。

## 消息参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `type` | 字符串 | 固定为 `ARTIFACT_UPLOADED`。 |
| `dispatchId` | 整数 | 本次派单 ID，同一工单重新派发时可能不同。 |
| `name` | 字符串 | 产物文件名。 |
| `artifactType` | 字符串 | 产物分类。 |
| `ossRef` | 字符串 | 产物在对象存储中的引用。 |
| `size` | 整数 | 文件大小，单位为字节。 |
| `metaJson` | 字符串 | 序列化为 JSON 字符串的产物元数据。 |

## 消息示例

```json
{
  "type": "ARTIFACT_UPLOADED",
  "dispatchId": 701,
  "name": "review.md",
  "artifactType": "REPORT",
  "ossRef": "example/review.md",
  "size": 2048
}
```

1.3.0 未为此消息定义单独的 WebSocket 回执。

## 相关协议

- [交付结果](../tasks/delivery.zh.md)

消息处理见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
