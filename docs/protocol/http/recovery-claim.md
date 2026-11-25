# Reclaim an active dispatch

English | [中文](recovery-claim.zh.md)

`POST /api/daemon/dispatches/{dispatchId}/recovery-claim`

After recovery, reclaim an active dispatch using its original dispatchId. Eligible states are DISPATCHED, ACKED and RUNNING.

## Request parameters

| Parameter | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| `dispatchId` | path | Integer | Yes | ID of this dispatch attempt for the work item. |
| `token` | query | String | Yes | Connection token of the executor assigned to this dispatch. |

## Example request

```sh
curl -X POST 'https://platform.example/api/daemon/dispatches/701/recovery-claim?token=EXAMPLE'
```

## Response fields

| Field | JSON type | Description |
| --- | --- | --- |
| `allowed` | Boolean | The executor is allowed to continue this dispatch. |
| `status` | String | Current dispatch state. |
| `environmentVariables` | Object | Environment variable names and string values for this execution. |

## Successful response

```json
{
  "allowed": true,
  "status": "RUNNING",
  "environmentVariables": {
    "LANG": "C.UTF-8"
  }
}
```

HTTP 404 indicates a missing dispatch, 409 a dispatch that cannot currently be reclaimed, and 401 an authentication or write-permission failure.

## Source

Handles the request and builds the response. [DaemonRecoveryClaimController.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DaemonRecoveryClaimController.java)
