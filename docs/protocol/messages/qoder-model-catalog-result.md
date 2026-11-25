# Return the model catalog

English | [中文](qoder-model-catalog-result.zh.md)

Executor → platform. Frame type: `QODER_MODEL_CATALOG_RESULT`.

Echo `requestId` and `provider`. On success send `success: true` and `models`. Each entry needs a nonempty `id` and `name`. The platform deduplicates IDs and treats an empty valid model list as failure.

## Message fields

| Field | JSON type | Description |
| --- | --- | --- |
| `type` | String | Fixed value `QODER_MODEL_CATALOG_RESULT`. |
| `requestId` | String | String identifier correlating the request with subsequent messages. |
| `provider` | String | Catalog provider, either qoder or qodercn. |
| `success` | Boolean | Whether the operation succeeded. |
| `models` | Array | Available models. Each entry needs nonempty id and name strings. |

## Example frame

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

## Related protocol

- [QODER_MODEL_CATALOG_REQUEST](qoder-model-catalog-request.md)
- [QODER_MODEL_CATALOG_V1](../features/qoder-model-catalog-v1.md)

Message handling in [ProviderModelCatalogService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ProviderModelCatalogService.java).
