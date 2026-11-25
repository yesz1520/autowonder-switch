# 查询模型目录

[English](qoder-model-catalog-request.md) | 中文

平台 → 执行器。帧类型为 `QODER_MODEL_CATALOG_REQUEST`。

按 `executorId` 指定执行器，按 `provider` 查询可用模型；`requestId` 标识本次查询。

## 消息参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `type` | 字符串 | 固定为 `QODER_MODEL_CATALOG_REQUEST`。 |
| `executorId` | 整数 | 该连接对应的执行器 ID。 |
| `requestId` | 字符串 | 关联请求与后续消息的字符串 ID。 |
| `provider` | 字符串 | 模型目录的提供方，支持 qoder 和 qodercn。 |

## 消息示例

```json
{
  "type": "QODER_MODEL_CATALOG_REQUEST",
  "executorId": 42,
  "requestId": "request-01",
  "provider": "qoder"
}
```

## 对应回包

回包参数见 [QODER_MODEL_CATALOG_RESULT](qoder-model-catalog-result.zh.md)。

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

- [QODER_MODEL_CATALOG_RESULT](qoder-model-catalog-result.zh.md)
- [QODER_MODEL_CATALOG_V1](../features/qoder-model-catalog-v1.zh.md)

消息处理见[ProviderModelCatalogService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ProviderModelCatalogService.java)。
