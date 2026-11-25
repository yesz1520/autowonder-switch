# Dispatch ownership inventory

English | [中文](dispatch-inventory-v1.zh.md)

Heartbeat capability identifier: `dispatch_inventory_v1`.

Include capacity, running and owned dispatches, and owned conversation turns in every heartbeat. Send the required arrays even when idle.

## Example declaration

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

## Related protocol

- [Execution inventory](../execution-inventory.md)

Capability implementation in [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java).
