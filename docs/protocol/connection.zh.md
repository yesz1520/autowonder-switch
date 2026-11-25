# 建立连接

[English](connection.md) | 中文

执行器用 `ws` 或 `wss` 连接平台的 `/ws/executor`。

建立连接时，客户端先发起 GET 请求，把 `executorId` 和 `token` 放在 URL 的查询字符串里，客户端在同一条 GET 请求中通过 `Upgrade: websocket` 请求头申请切换协议。平台返回 HTTP 101 后校验执行器身份。如果执行器不存在、已删除或令牌不匹配，平台会发送关闭码为 `1008`、原因为 `authentication failed` 的 WebSocket 关闭帧。

## 请求参数

| 查询参数 | 必填 | 描述 | 服务端行为 |
|---|---|---|---|
| `executorId` | 是 | 发起连接的执行器 ID。 | 按十进制有符号 64 位整数解析，再查找该 ID 对应的执行器。 |
| `token` | 是 | 该执行器的身份凭据。 | 校验该执行器的令牌。认证成功后，平台从执行器记录取得租户和智能体身份。 |
| `hostname` | 否 | 执行器所在机器的主机名。 | 记录最近接入的主机名。去除首尾空白和控制字符，最多保留 128 个字符；空视为未提供。 |
| `maxConcurrentDispatches` | 否 | 该执行器声明的并发容量，供平台判断是否还能派发任务。 | 握手时省略取 3；1–50 按原值使用；可解析的整数大于 50 取 50；非正数、空串或无法解析为 32 位整数的值取 1。 |

## 握手示例

```http
GET /ws/executor?executorId=42&token=EXAMPLE HTTP/1.1
Host: platform.example
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Sec-WebSocket-Version: 13
```

服务器接受协议升级时返回以下响应。执行器身份校验失败时的关闭帧见下文的关闭码表。

```http
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

## 接入 IP

接入 IP 来自升级请求的 HTTP 头。平台依次查找 `X-Forwarded-For` 和 `X-Real-IP`，取第一个非空且不等于 `unknown` 的值。两者都没有可用值时，本次连接解析不到 IP。握手中无需传 `ip` 查询参数。

## 关闭码

| 关闭码 | 原因文本 | 触发条件 |
|---|---|---|
| 1008 | `missing token or executorId` | 缺少任一必填查询参数。 |
| 1008 | `invalid executorId` | ID 无法解析为 64 位整数。 |
| 1008 | `authentication failed` | 执行器不存在、已删除，或令牌不匹配。 |
| 1008 | `session initialization busy` | 认证通过，但平台未能登记本次会话。 |
| 4001 | `executor connection replaced` | 同一服务节点上，该 ID 的新连接替换了旧连接。关闭帧发给旧连接。 |

连接建立后，执行器发送[心跳](heartbeat.zh.md)。`startedAt`、`version` 和 `model` 随心跳上报。

## 相关源码

- 解析连接参数并发送关闭帧。 [ExecutorWsEndpoint.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/ExecutorWsEndpoint.java)
- 校验执行器 ID 和令牌。 [ExecutorWsAuthenticator.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/ExecutorWsAuthenticator.java)
- 登记执行器当前的会话 ID。 [PresenceManager.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/PresenceManager.java)
- 同一执行器重新连接时，关闭本节点上的旧连接。 [SessionRegistry.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/SessionRegistry.java)
- 从握手请求头读取接入 IP。 [ClientIpResolver.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/ClientIpResolver.java)
- 清理主机名并记录最近接入的 IP 和主机名。 [ExecutorService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorService.java)
