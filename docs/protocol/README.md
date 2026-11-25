# AutoWonder executor protocol

English | [中文](README.zh.md)

The executor opens a WebSocket to AutoWonder. The platform sends dispatches over this connection; the executor acknowledges them, reports progress and returns results. Task packages are downloaded via signed URLs. Artifacts and checkpoints are uploaded over HTTP.

Version: [Community 1.3.0](https://github.com/aliyun/alibabacloud-autowonder/releases/tag/1.3.0)

Date: November 23, 2026

Commit: [0a58555](https://github.com/aliyun/alibabacloud-autowonder/commit/0a585550f1d2369f73b2b8b620c206d8f5ec1e57)

## Follow one task

1. [Establish a connection](connection.md). Check the authentication parameters, client IP and close codes.
2. Send a [heartbeat](heartbeat.md) and report capacity, dispatches and conversation turns in the [execution inventory](execution-inventory.md).
3. [Accept the dispatch](tasks/dispatch.md). Read the ZIP, select the model and report progress.
4. [Deliver the result](tasks/delivery.md). Upload the output and completion checkpoint, then match the result receipt.
5. [Pause or recover](tasks/recovery.md). Decide what to do after a disconnect or process restart.

## Other exchanges

- [Conversation turns](conversations.md) use turn IDs and their own delivery receipts.
- [Capabilities and controls](capabilities.md) covers model discovery, guidance, handoff and remote lifecycle commands.

## Dispatch HTTP endpoints

- [Refresh a package URL](http/package-url.md)
- [Upload dispatch artifacts](http/artifacts.md)
- [Upload a checkpoint](http/checkpoint.md)
- [Refresh a checkpoint URL](http/checkpoint-url.md)
- [Reclaim an active dispatch](http/recovery-claim.md)
