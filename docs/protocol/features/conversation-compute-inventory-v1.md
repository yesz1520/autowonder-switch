# Computing turn inventory

English | [中文](conversation-compute-inventory-v1.zh.md)

Heartbeat capability identifier: `CONVERSATION_COMPUTE_INVENTORY_V1`.

Report computing turns in the heartbeat’s `computingConversationTurnIds`. Each ID must belong to an owned turn. Send an empty array when none are computing.

## Example declaration

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

## Related protocol

- [Execution inventory](../execution-inventory.md)

Capability implementation in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
