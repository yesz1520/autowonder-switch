# Using the Go protocol client

English | [中文](protocol-client.zh.md)

A daemon in this repository talks to the platform through `internal/autowonder`. Each `Client` represents one executor ID on one platform.

## Connect an identity

Pass the platform HTTP(S) base URL, executor ID and token to `New` in a `Config`. Call `OpenOutbox` once for this client, then pass it and a heartbeat snapshot function to `NewEndpoint`. Identities can share a state directory. Files are scoped by platform and executor ID.

Register incoming frame handlers with `On`, advertise supported capabilities in the heartbeat, then call `client.Run`. It sends an immediate heartbeat and reconnects after network loss. Authentication failure or replacement by another connection returns an error. For model catalog wiring see [the Go example](../../internal/autowonder/example_test.go).

## Accept a dispatch and deliver its result

On `TaskDispatch`, check whether the executor can accept the task. Acknowledge an accepted dispatch with `TaskAck`, then download and verify its package in a worker before running it. Return from the callback promptly so pause commands, guidance and receipts can be received. Use an executor-owned context for the task. The callback context ends at disconnect.

Generate and persist a `NewResultID()` for a V2 result. On success, upload a completion checkpoint with that ID and copy its returned sequence and digest into `TaskResult`. Deliver it through `endpoint.Send` to persist the frame before writing it to the socket. A nil session queues the delivery while offline. A matching terminal receipt clears it. A retryable receipt leaves it pending. Call `outbox.Replay` to retry on the current connection.

## Call the HTTP endpoints

Use the same `Client` for package downloads, checkpoints, artifact and diagnostic uploads, task data, human requests, usage and environment snapshots. For memory operations call `client.Memory(dispatchID, dispatchMCPToken)` with the dispatch credential. Use `errors.As` with `*HTTPError` for the status code and error body. Check individual artifact receipts even when the HTTP request succeeds.

For directions, fields and examples see [the executor protocol](../protocol/README.md).
