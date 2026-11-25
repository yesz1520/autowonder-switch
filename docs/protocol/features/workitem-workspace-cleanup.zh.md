# 工单工作区清理

[English](workitem-workspace-cleanup.md) | 中文

心跳能力标识为 `WORKITEM_WORKSPACE_CLEANUP`。

支持按工单版本清理工作区。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "WORKITEM_WORKSPACE_CLEANUP"
  ]
}
```

## 相关协议

- [WORKITEM_WORKSPACE_CLEANUP](../messages/workitem-workspace-cleanup.zh.md)

能力定义与使用见[WorkspaceCleanupTask.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/workspace/WorkspaceCleanupTask.java)、[WsWorkspaceCleanupTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/workspace/WsWorkspaceCleanupTransport.java)。
