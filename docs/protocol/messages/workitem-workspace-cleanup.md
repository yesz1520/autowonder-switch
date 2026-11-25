# Clean up a workspace

English | [中文](workitem-workspace-cleanup.zh.md)

Platform → executor. Frame type: `WORKITEM_WORKSPACE_CLEANUP`.

The platform requests cleanup for a closed work item on this executor. It uses `workitemVersion` to distinguish cleanup requests for different versions of the same work item.

## Message fields

| Field | JSON type | Description |
| --- | --- | --- |
| `type` | String | Fixed value `WORKITEM_WORKSPACE_CLEANUP`. |
| `executorId` | Integer | ID of the executor on this connection. |
| `tenantId` | Integer | Tenant that owns the work item. |
| `workitemId` | Integer | ID of the work item. |
| `workitemVersion` | Integer | Work item version to clean up. |
| `publishedAt` | Integer | Time of the last transition into the current terminal status, in Unix milliseconds. |

## Example frame

```json
{
  "type": "WORKITEM_WORKSPACE_CLEANUP",
  "executorId": 42,
  "tenantId": 3,
  "workitemId": 91,
  "workitemVersion": 3,
  "publishedAt": 1795593600000
}
```

Version 1.3.0 defines no dedicated WebSocket receipt for this message.

## Related protocol

- [WORKITEM_WORKSPACE_CLEANUP](../features/workitem-workspace-cleanup.md)

Message handling in [WsWorkspaceCleanupTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/workspace/WsWorkspaceCleanupTransport.java).
