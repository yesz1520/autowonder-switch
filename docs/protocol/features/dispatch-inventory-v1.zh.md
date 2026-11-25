# 派单持有清单

[English](dispatch-inventory-v1.md) | 中文

心跳能力标识为 `dispatch_inventory_v1`。

在每次心跳中列出容量、正在执行和仍持有的派单，以及持有的会话轮次。空闲时也要发送必填数组。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "dispatch_inventory_v1"
  ],
  "maxConcurrentDispatches": 2,
  "dispatchInventoryReady": true,
  "runningDispatchIds": [],
  "ownedDispatchIds": [],
  "runningConversationTurnIds": []
}
```

## 相关协议

- [执行清单](../execution-inventory.zh.md)

能力定义与使用见[InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)。
