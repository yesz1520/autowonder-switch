# 轮次计算清单

[English](conversation-compute-inventory-v1.md) | 中文

心跳能力标识为 `CONVERSATION_COMPUTE_INVENTORY_V1`。

在心跳的 `computingConversationTurnIds` 中列出正在计算的轮次。每个 ID 都必须属于当前持有的轮次，当前没有计算时传空数组。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "dispatch_inventory_v1",
    "CONVERSATION_COMPUTE_INVENTORY_V1"
  ],
  "maxConcurrentDispatches": 2,
  "dispatchInventoryReady": true,
  "runningDispatchIds": [],
  "ownedDispatchIds": [],
  "runningConversationTurnIds": [],
  "computingConversationTurnIds": []
}
```

## 相关协议

- [执行清单](../execution-inventory.zh.md)

能力定义与使用见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
