# 返回模型目录

[English](qoder-model-catalog-result.md) | 中文

执行器 → 平台。帧类型为 `QODER_MODEL_CATALOG_RESULT`。

原样回传 `requestId` 和 `provider`。成功时填 `success: true` 和 `models`，每项必须包含非空的 `id` 和 `name`。平台按 ID 去重，没有有效模型视为失败。

## 消息参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `type` | 字符串 | 固定为 `QODER_MODEL_CATALOG_RESULT`。 |
| `requestId` | 字符串 | 关联请求与后续消息的字符串 ID。 |
| `provider` | 字符串 | 模型目录的提供方，支持 qoder 和 qodercn。 |
| `success` | 布尔值 | 操作是否成功。 |
| `models` | 数组 | 可用模型列表，每项的 id 和 name 均为非空字符串。 |

## 消息示例

```json
{
  "type": "QODER_MODEL_CATALOG_RESULT",
  "requestId": "request-01",
  "provider": "qoder",
  "success": true,
  "models": [
    {
      "id": "example-model",
      "name": "Example Model"
    }
  ]
}
```

## 相关协议

- [QODER_MODEL_CATALOG_REQUEST](qoder-model-catalog-request.zh.md)
- [QODER_MODEL_CATALOG_V1](../features/qoder-model-catalog-v1.zh.md)

消息处理见[ProviderModelCatalogService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ProviderModelCatalogService.java)。
