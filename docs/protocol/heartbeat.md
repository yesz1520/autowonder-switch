# Heartbeat

English | [中文](heartbeat.zh.md)

Once connected, the executor sends a JSON text frame over the WebSocket and repeats it periodically. Each successful heartbeat renews online presence for 90 seconds. Send heartbeats less than 90 seconds apart.

## Basic fields

| Field | JSON type | When to send | Description |
|---|---|---|---|
| `type` | String | Always | Use `HEARTBEAT`. |
| `startedAt` | String | Optional | Executor process start time in ISO 8601 format with a UTC offset, for example `2026-11-25T08:00:00Z`. |
| `version` | String | When reporting a version | Version of the running executor. The platform trims surrounding whitespace and keeps up to 64 UTF-16 code units. |
| `model` | String | When reporting a model | ID of the model the executor currently uses. The platform trims surrounding whitespace and keeps up to 128 UTF-16 code units. |
| `protocolFeatures` | Array of strings | Every heartbeat that advertises capabilities | Send the full list of supported [protocol features](capabilities.md). Names are case-sensitive. The platform reads the first 64 entries. Omitting the field or sending `[]` advertises no capabilities for this heartbeat. |
| `restartRequestId` | String | When confirming a platform-requested restart | Echo the restart command’s `requestId` along with the new process’s `startedAt`. If the platform specified a target version, report the matching `version` as well. |

Each of `version` and `model` expires 90 seconds after its last nonblank report. Omitting a field or sending a blank string leaves its previous expiry in place. Include both current values in every heartbeat to keep them visible on the platform.

## Example heartbeat

```json
{
  "type": "HEARTBEAT",
  "startedAt": "2026-11-25T08:00:00Z",
  "version": "1.2.3",
  "model": "example-model"
}
```

## Execution inventory

When advertising `dispatch_inventory_v1`, include the inventory in every heartbeat, even when idle. See [Execution inventory](execution-inventory.md) for the fields and an example.

## Source files

- Parses heartbeat fields and validates inventory. [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)
- Records capabilities, version, model and presence expiry. [PresenceManager.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/PresenceManager.java)
- Validates startup time and confirms restart requests. [ExecutorRestartService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorRestartService.java)
