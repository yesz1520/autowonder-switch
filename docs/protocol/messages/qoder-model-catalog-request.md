# Request the model catalog

English | [中文](qoder-model-catalog-request.zh.md)

Platform → executor. Frame type: `QODER_MODEL_CATALOG_REQUEST`.

The platform selects the executor with `executorId` and requests models for `provider`. The `requestId` identifies this query.

## Message fields

| Field | JSON type | Description |
| --- | --- | --- |
| `type` | String | Fixed value `QODER_MODEL_CATALOG_REQUEST`. |
| `executorId` | Integer | ID of the executor on this connection. |
| `requestId` | String | String identifier correlating the request with subsequent messages. |
| `provider` | String | Catalog provider, either qoder or qodercn. |

## Example frame

```json
{
  "type": "QODER_MODEL_CATALOG_REQUEST",
  "executorId": 42,
  "requestId": "request-01",
  "provider": "qoder"
}
```

## Example reply

See [QODER_MODEL_CATALOG_RESULT](qoder-model-catalog-result.md) for the reply fields.

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

- [QODER_MODEL_CATALOG_RESULT](qoder-model-catalog-result.md)
- [QODER_MODEL_CATALOG_V1](../features/qoder-model-catalog-v1.md)

Message handling in [ProviderModelCatalogService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ProviderModelCatalogService.java).
