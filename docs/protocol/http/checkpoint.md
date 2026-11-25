# Upload a checkpoint

English | [中文](checkpoint.zh.md)

`POST /api/daemon/dispatches/{dispatchId}/checkpoint`

Save the archive to checkpointUploadPath. This example uploads a completion checkpoint. Use the same resultId in the subsequent task result.

## Request parameters

| Parameter | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `dispatchId` | path | Integer | Yes | ID of this dispatch attempt for the work item. |
| `token` | query | String | Yes | Connection token of the executor assigned to this dispatch. |
| `checkpointSeq` | multipart | Integer | Yes | Positive checkpoint sequence number for the dispatch. |
| `checkpoint` | multipart | File | Yes | Nonempty checkpoint archive file. |
| `purpose` | multipart | String | Optional | Use COMPLETION for a completed task. For a recovery checkpoint use RECOVERY or omit. |
| `resultId` | multipart | String | Conditional | Required for COMPLETION. Omit for a recovery checkpoint. |
| `provider` | multipart | String | Optional | Provider that produced this checkpoint. |
| `providerSessionId` | multipart | String | Optional | Native session ID associated with the checkpoint. |
| `runtimeId` | multipart | String | Optional | Identifier of the runtime producing the checkpoint. |
| `activeStepId` | multipart | String | Optional | Active step when the checkpoint was saved. |

## Example request

```sh
curl -X POST 'https://platform.example/api/daemon/dispatches/701/checkpoint?token=EXAMPLE' -F 'checkpointSeq=2' -F 'checkpoint=@checkpoint.zip' -F 'purpose=COMPLETION' -F 'resultId=86ebc809-2d32-44e4-b456-29b870f878bd'
```

## Response fields

| Field | JSON type | Description |
| --- | --- | --- |
| `checkpointSeq` | Integer | Positive checkpoint sequence number for the dispatch. |
| `sha256` | String | SHA-256 digest of the checkpoint archive. |
| `sizeBytes` | Integer | Size of the uploaded file in bytes. |
| `purpose` | String | Use COMPLETION for a completed task. For a recovery checkpoint use RECOVERY or omit. Returned only for completion checkpoints. |
| `resultId` | String | Lowercase UUID identifying the result across retries. Returned only for completion checkpoints. |

## Successful response

```json
{
  "checkpointSeq": 2,
  "sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "sizeBytes": 2048,
  "purpose": "COMPLETION",
  "resultId": "86ebc809-2d32-44e4-b456-29b870f878bd"
}
```

The checkpoint file limit is 50 MiB. An empty or oversized file, or a nonpositive sequence number, yields HTTP 400.

Scalar form parameters can also be sent in the query string. Send the file as the multipart `checkpoint` part. A retry with the same dispatch and `checkpointSeq` must contain identical bytes or the server returns HTTP 409. To include checkpoint metadata, put `checkpoint.json` at the root of a tar.gz archive. [DispatchCheckpointService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DispatchCheckpointService.java)

## Source

Handles the request and builds the response. [DaemonCheckpointController.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DaemonCheckpointController.java)
