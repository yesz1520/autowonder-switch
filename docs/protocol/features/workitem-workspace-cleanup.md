# Work item workspace cleanup

English | [中文](workitem-workspace-cleanup.zh.md)

Heartbeat capability identifier: `WORKITEM_WORKSPACE_CLEANUP`.

Supports workspace cleanup by work item version.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "WORKITEM_WORKSPACE_CLEANUP"
  ]
}
```

## Related protocol

- [WORKITEM_WORKSPACE_CLEANUP](../messages/workitem-workspace-cleanup.md)

Capability implementation in [WorkspaceCleanupTask.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/workspace/WorkspaceCleanupTask.java) and [WsWorkspaceCleanupTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/workspace/WsWorkspaceCleanupTransport.java).
