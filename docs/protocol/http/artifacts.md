# Upload dispatch artifacts

English | [中文](artifacts.zh.md)

`POST /api/daemon/dispatches/{dispatchId}/artifacts`

Upload files to artifactUploadPath. Check each entry in the response files array, including on HTTP 200.

## Request parameters

| Parameter | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `dispatchId` | path | Integer | Yes | ID of this dispatch attempt for the work item. |
| `token` | query | String | Yes | Connection token of the executor assigned to this dispatch. |
| `files` | multipart | File | Yes | Multipart file part. Repeat the part to upload multiple files. |
| `filesMetadata` | multipart | String | Optional | JSON array string with metadata such as paths for the uploaded files. |
| `idempotencyKey` | multipart | String | Optional | Stable key for retries of the same operation. |

## Example request

```sh
curl -X POST 'https://platform.example/api/daemon/dispatches/701/artifacts?token=EXAMPLE' -F 'files=@review.md'
```

## Response fields

| Field | JSON type | Description |
| --- | --- | --- |
| `remoteRef` | String | Storage prefix for this upload. |
| `files` | Array | Per-file receipts. index is zero-based, path is the file path and status is ACCEPTED or REJECTED. Accepted files include storage references and sizeBytes. Rejections include code and may include maxBytes. |

## Successful response

```json
{
  "remoteRef": "example/701/",
  "files": [
    {
      "index": 0,
      "path": "review.md",
      "status": "ACCEPTED",
      "sizeBytes": 2048,
      "ossRef": "example/701/review.md",
      "remoteRef": "example/701/review.md"
    }
  ]
}
```

A request accepts up to 200 files. Exceeding that limit yields HTTP 400.

Each file can be up to 50 MiB. Send `filesMetadata` as a multipart text part containing a JSON array string. Its indices match the order of the `files` parts. If an entry has no `path`, the platform uses the multipart filename. In this receipt excerpt the second file was rejected. Its `code` gives the reason.

```json
{
  "files": [
    {
      "index": 0,
      "path": "review.md",
      "status": "ACCEPTED",
      "sizeBytes": 200,
      "ossRef": "example/review.md",
      "remoteRef": "example/review.md"
    },
    {
      "index": 1,
      "path": "large.bin",
      "status": "REJECTED",
      "code": "FILE_TOO_LARGE",
      "sizeBytes": 52428801,
      "maxBytes": 52428800
    }
  ]
}
```

## Source

Handles the request and builds the response. [DaemonArtifactController.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/artifact/DaemonArtifactController.java)
