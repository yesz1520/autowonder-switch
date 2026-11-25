# 清理指定工作区

[English](workitem-workspace-cleanup.md) | 中文

平台 → 执行器。帧类型为 `WORKITEM_WORKSPACE_CLEANUP`。

平台请求清理该执行器上已结束工单的工作区。同一工单不同版本的清理请求通过 `workitemVersion` 区分。

## 消息参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `type` | 字符串 | 固定为 `WORKITEM_WORKSPACE_CLEANUP`。 |
| `executorId` | 整数 | 该连接对应的执行器 ID。 |
| `tenantId` | 整数 | 工单所属租户的 ID。 |
| `workitemId` | 整数 | 派单所属工单的 ID。 |
| `workitemVersion` | 整数 | 要清理的工单版本。 |
| `publishedAt` | 整数 | 工单最近一次进入当前终态的时间，Unix 毫秒时间戳。 |

## 消息示例

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

1.3.0 未为此消息定义单独的 WebSocket 回执。

## 相关协议

- [WORKITEM_WORKSPACE_CLEANUP](../features/workitem-workspace-cleanup.zh.md)

消息处理见[WsWorkspaceCleanupTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/workspace/WsWorkspaceCleanupTransport.java)。
