# Available model catalog

English | [中文](qoder-model-catalog-v1.zh.md)

Heartbeat capability identifier: `QODER_MODEL_CATALOG_V1`.

Supports querying models by provider. Version 1.3.0 supports `qoder` and `qodercn`. Echo the provider and request ID in the response.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "QODER_MODEL_CATALOG_V1"
  ]
}
```

## Related protocol

- [QODER_MODEL_CATALOG_REQUEST](../messages/qoder-model-catalog-request.md)
- [QODER_MODEL_CATALOG_RESULT](../messages/qoder-model-catalog-result.md)

Capability implementation in [ProviderModelCatalogService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ProviderModelCatalogService.java).
