# Register an uploaded artifact

English | [中文](artifact-uploaded.zh.md)

Executor → platform. Frame type: `ARTIFACT_UPLOADED`.

Report an uploaded artifact. Identify the dispatch by `dispatchId` and describe the artifact with `name`, `artifactType`, `ossRef`, `size` and `metaJson`.

## Message fields

| Field | JSON type | Description |
| --- | --- | --- |
| `type` | String | Fixed value `ARTIFACT_UPLOADED`. |
| `dispatchId` | Integer | ID of this dispatch attempt for the work item. |
| `name` | String | Artifact filename. |
| `artifactType` | String | Artifact category. |
| `ossRef` | String | Object storage reference for the artifact. |
| `size` | Integer | File size in bytes. |
| `metaJson` | String | Artifact metadata serialized as a JSON string. |

## Example frame

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

Version 1.3.0 defines no dedicated WebSocket receipt for this message.

## Related protocol

- [Deliver results](../tasks/delivery.md)

Message handling in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
