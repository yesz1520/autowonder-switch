# 心跳

[English](heartbeat.md) | 中文

连接建立后，执行器通过 WebSocket 发送 JSON 文本帧，每次成功上报都会刷新在线状态，有效期为 90 秒，故发送间隔应短于 90 秒。

## 基本字段

| 字段 | JSON 类型 | 何时发送 | 描述 |
|---|---|---|---|
| `type` | 字符串 | 每次 | 固定为 `HEARTBEAT`。 |
| `startedAt` | 字符串 | 可选 | 执行器进程的启动时间，使用带时区偏移的 ISO 8601 格式，例如 `2026-11-25T08:00:00Z`。 |
| `version` | 字符串 | 上报版本时 | 当前运行的执行器版本。平台去除首尾空白，最多保留 64 个 UTF-16 码元。 |
| `model` | 字符串 | 上报模型时 | 执行器当前使用的模型 ID。平台去除首尾空白，最多保留 128 个 UTF-16 码元。 |
| `protocolFeatures` | 字符串数组 | 声明能力时，每次发送 | 填写当前支持的全部[协议能力](capabilities.zh.md)，名称区分大小写。平台读取前 64 项。省略或传 `[]` 表示本次没有声明能力。 |
| `restartRequestId` | 字符串 | 确认平台发起的重启时 | 原样回传重启请求的 `requestId`，同时上报新进程的 `startedAt`。若平台指定了目标版本，还需在 `version` 中上报匹配的版本。 |

> `version` 和 `model` 从最后一次收到非空值起保留 90 秒。

## 心跳示例

```json
{
  "type": "HEARTBEAT",
  "startedAt": "2026-11-25T08:00:00Z",
  "version": "1.2.3",
  "model": "example-model"
}
```

## 执行清单

声明 `dispatch_inventory_v1` 后，每次心跳都要携带执行清单，空闲时也要发送。字段和示例见[执行清单](execution-inventory.zh.md)。

## 相关源码

- 解析心跳字段并校验清单。 [InboundFrameRouter.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/InboundFrameRouter.java)
- 记录能力、版本、模型和在线状态的有效期。 [PresenceManager.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/PresenceManager.java)
- 校验启动时间并确认重启请求。 [ExecutorRestartService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorRestartService.java)
