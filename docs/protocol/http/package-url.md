# Refresh a package URL

English | [中文](package-url.zh.md)

`POST /api/daemon/dispatches/{dispatchId}/package-url`

When the package URL expires, use the dispatch’s packageRefreshPath to get a fresh link.

## Request parameters

| Parameter | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `dispatchId` | path | Integer | Yes | ID of this dispatch attempt for the work item. |
| `token` | query | String | Yes | Connection token of the executor assigned to this dispatch. |

## Example request

```sh
curl -X POST 'https://platform.example/api/daemon/dispatches/701/package-url?token=EXAMPLE'
```

## Response fields

| Field | JSON type | Description |
| --- | --- | --- |
| `downloadUrl` | String | Signed URL for downloading the file. |
| `expiresInSeconds` | Integer | Seconds until the newly signed URL expires. |

## Successful response

```json
{
  "downloadUrl": "https://storage.example/task.zip",
  "expiresInSeconds": 900
}
```

HTTP 409 means the package URL cannot be renewed for this dispatch.

## Source

Handles the request and builds the response. [DaemonTaskPackageController.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/taskpackage/DaemonTaskPackageController.java)
