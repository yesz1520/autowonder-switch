# 可用模型目录

[English](qoder-model-catalog-v1.md) | 中文

心跳能力标识为 `QODER_MODEL_CATALOG_V1`。

支持按 provider 查询可用模型。1.3.0 支持 `qoder` 和 `qodercn`，响应要带回请求中的 provider 和请求 ID。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "QODER_MODEL_CATALOG_V1"
  ]
}
```

## 相关协议

- [QODER_MODEL_CATALOG_REQUEST](../messages/qoder-model-catalog-request.zh.md)
- [QODER_MODEL_CATALOG_RESULT](../messages/qoder-model-catalog-result.zh.md)

能力定义与使用见[ProviderModelCatalogService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ProviderModelCatalogService.java)。
