# Refresh a checkpoint URL

English | [中文](checkpoint-url.zh.md)

`POST /api/daemon/dispatches/{dispatchId}/checkpoint-url`

Refresh a selected checkpoint link using the dispatch’s checkpointRefreshPath and the chosen sequence number.

## Request parameters

| Parameter | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `dispatchId` | path | Integer | Yes | ID of this dispatch attempt for the work item. |
| `token` | query | String | Yes | Connection token of the executor assigned to this dispatch. |
| `checkpointSeq` | query | Integer | Yes | Positive checkpoint sequence number for the dispatch. |

## Example request

```sh
curl -X POST 'https://platform.example/api/daemon/dispatches/701/checkpoint-url?token=EXAMPLE&checkpointSeq=2'
```

## Response fields

| Field | JSON type | Description |
| --- | --- | --- |
| `checkpointSeq` | Integer | Positive checkpoint sequence number for the dispatch. |
| `downloadUrl` | String | Signed URL for downloading the file. |
| `expiresAt` | String | Expiry timestamp. |
| `expiresInSeconds` | Integer | Seconds until the newly signed URL expires. |
| `sha256` | String | SHA-256 digest of the checkpoint archive. |

## Successful response

```json
{
  "checkpointSeq": 2,
  "downloadUrl": "https://storage.example/checkpoint.zip",
  "expiresAt": "2026-11-25T09:00:00Z",
  "expiresInSeconds": 900,
  "sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
```

## Source

Handles the request and builds the response. [DaemonCheckpointController.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DaemonCheckpointController.java)
