# Establish a connection

English | [中文](connection.zh.md)

The executor connects to `/ws/executor` over `ws` or `wss`.

To open the connection, the client sends an HTTP GET request with `executorId` and `token` in the URL query string. The `Upgrade: websocket` header in that same GET request asks the server to switch protocols. After returning HTTP 101, the platform authenticates the executor. If the executor does not exist, has been deleted, or its token does not match, the platform sends a WebSocket close frame with code `1008` and reason `authentication failed`.

## Query parameters

| Parameter | Required | Description | Server behavior |
|---|---|---|---|
| `executorId` | Yes | ID of the connecting executor. | Parses a signed 64-bit decimal integer and looks up the executor by ID. |
| `token` | Yes | Authentication credential for this executor. | Validates the token for that executor. After authentication, the platform takes the tenant and agent identities from the executor record. |
| `hostname` | No | Name of the machine hosting the executor. | Records the last connecting hostname. Trims whitespace, removes control characters and keeps up to 128 characters. An empty result is treated as absent. |
| `maxConcurrentDispatches` | No | Concurrency capacity advertised by this executor, used to decide whether it can receive more work. | Defaults to 3 when omitted at connection time. Values 1–50 are kept. Larger parsed integers are capped at 50. Nonpositive values, an empty string and values that cannot be parsed as 32-bit integers yield 1. |

## Example handshake

```http
GET /ws/executor?executorId=42&token=EXAMPLE HTTP/1.1
Host: platform.example
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Sec-WebSocket-Version: 13
```

The server returns this response when it accepts the protocol upgrade. Authentication failures use the close codes listed below.

```http
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

## Client IP

The platform reads the IP from the HTTP upgrade headers. It checks `X-Forwarded-For` first, then `X-Real-IP`, taking the first nonempty value other than `unknown`. If neither header supplies a usable value, no IP is resolved for this connection. There is no `ip` query parameter to supply.

## Close codes

| Code | Reason text | Condition |
|---|---|---|
| 1008 | `missing token or executorId` | A required query parameter is absent. |
| 1008 | `invalid executorId` | The ID cannot be parsed as a 64-bit integer. |
| 1008 | `authentication failed` | The executor does not exist, has been deleted or the token does not match. |
| 1008 | `session initialization busy` | Authentication succeeded but the platform could not announce the session. |
| 4001 | `executor connection replaced` | A new connection for the same ID replaced the old one on the same server node. The old socket receives the close frame. |

Once connected, the executor sends its [heartbeat](heartbeat.md). The heartbeat carries `startedAt`, `version` and `model`.

## Source files

- Parses connection parameters and sends close frames. [ExecutorWsEndpoint.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/ExecutorWsEndpoint.java)
- Validates the executor ID and token. [ExecutorWsAuthenticator.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/ExecutorWsAuthenticator.java)
- Records the current session ID for the executor. [PresenceManager.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/PresenceManager.java)
- Closes the previous connection on the same node when an executor reconnects. [SessionRegistry.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/SessionRegistry.java)
- Reads the client IP from the handshake headers. [ClientIpResolver.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/ClientIpResolver.java)
- Sanitizes the hostname and records the last connection’s IP and hostname. [ExecutorService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorService.java)
