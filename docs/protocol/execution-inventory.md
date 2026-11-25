# Execution inventory

English | [中文](execution-inventory.zh.md)

The executor reports its capacity, dispatches and conversation turns in the [heartbeat](heartbeat.md). Inventory fields appear at the same level as `type`.

## Inventory fields

When advertising `dispatch_inventory_v1`, send all five required fields below in every heartbeat, including while idle or while the inventory is incomplete. Encode IDs as positive JSON integers without quotes. Each ID may occur only once per array. If a required field is missing, has the wrong type, or violates an inventory constraint, the platform records a protocol error and closes the connection.

| Field | JSON type | When to send | Description |
|---|---|---|---|
| `maxConcurrentDispatches` | Integer | Required | Total concurrency capacity, from 1 to 50. Include it even if the executor supplied it in the handshake. |
| `dispatchInventoryReady` | Boolean | Required | Send `true` when the inventory is complete and the platform can reconcile it. Send `false` while restoring local state or when the executor cannot list all dispatches it owns. |
| `runningDispatchIds` | Array of integers | Required | IDs of dispatches currently executing, up to 50. Send `[]` when none are running. |
| `ownedDispatchIds` | Array of integers | Required | IDs of dispatches still owned by the executor, up to 1,000. Include every ID in `runningDispatchIds` and any finished dispatch still awaiting its result receipt. Send `[]` when none are owned. |
| `runningConversationTurnIds` | Array of integers | Required | The `turnId` of each conversation turn the executor still owns, up to 50. Send `[]` when none are owned. |
| `dispatchInventoryError` | String | When the executor owns more than 1,000 dispatches | Send `OWNED_DISPATCH_LIMIT_EXCEEDED`, set `dispatchInventoryReady` to `false`, and send only the running dispatch IDs in both dispatch arrays. Normally omit this field. |
| `computingConversationTurnIds` | Array of integers | When advertising `CONVERSATION_COMPUTE_INVENTORY_V1` | IDs of owned turns currently computing, up to 50. Send `[]` when none are computing. If the field is missing, malformed, or includes an unowned turn, the platform treats compute activity as unknown and can still use the rest of a valid heartbeat. |

Without `dispatch_inventory_v1`, the platform uses the handshake capacity. The executor can still report `runningDispatchIds` and `runningConversationTurnIds`. Older clients can also send `runningConversationTurns` as an array of objects with a `turnId` field. Omitting both turn fields leaves turn activity unreported. An explicit empty array reports that the executor owns no turns.

## Example heartbeat

Dispatch 701 has finished executing and is awaiting its result receipt, so it appears only in `ownedDispatchIds`.

```json
{
  "type": "HEARTBEAT",
  "startedAt": "2026-11-25T08:00:00Z",
  "version": "1.2.3",
  "model": "example-model",
  "protocolFeatures": [
    "dispatch_inventory_v1",
    "CONVERSATION_COMPUTE_INVENTORY_V1"
  ],
  "maxConcurrentDispatches": 2,
  "dispatchInventoryReady": true,
  "runningDispatchIds": [],
  "ownedDispatchIds": [
    701
  ],
  "runningConversationTurnIds": [],
  "computingConversationTurnIds": []
}
```

## Source files

- Parses heartbeat fields and validates inventory. [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)
