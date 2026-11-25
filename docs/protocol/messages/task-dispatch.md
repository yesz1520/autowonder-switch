# Dispatch a task

English | [中文](task-dispatch.zh.md)

Platform → executor. Frame type: `TASK_DISPATCH`.

The platform identifies the dispatch by `dispatchId` and its recipient by `executorId`. If the executor accepts the dispatch, send `TASK_ACK` before fetching and verifying the package from `downloadUrl`. Otherwise report `TASK_BUSY`.

## Message fields

| Field | JSON type | Condition | Description |
| --- | --- | --- | --- |
| `type` | String |  | Fixed value `TASK_DISPATCH`. |
| `executorId` | Integer |  | ID of the executor on this connection. |
| `dispatchId` | Integer |  | ID of this dispatch attempt for the work item. |
| `tenantId` | Integer |  | Tenant that owns the work item. |
| `workitemId` | Integer |  | ID of the work item. |
| `agentId` | Integer |  | Agent assigned to this execution. |
| `agentVersionId` | Integer |  | Agent version used for this execution. |
| `attempt` | Integer |  | Dispatch attempt number. |
| `downloadUrl` | String |  | Signed URL for downloading the file. |
| `size` | Integer |  | File size in bytes. |
| `checksumAlgorithm` | String |  | Algorithm used for the package digest. |
| `checksumScope` | String |  | Digest scope. This value covers the complete ZIP. |
| `checksum` | String |  | SHA-256 digest of the package. |
| `artifactUploadPath` | String |  | HTTP path for artifacts. |
| `checkpointUploadPath` | String |  | HTTP path for checkpoint. |
| `packageRefreshPath` | String |  | HTTP path for package-url. |
| `sdlcStepId` | Integer |  | Workflow step associated with the dispatch. |
| `idempotencyKey` | String |  | Stable key for retries of the same operation. |
| `md5` | String |  | Hexadecimal MD5 digest of the task archive. |
| `packageId` | String |  | Task package identifier. |
| `issuer` | String | Signed package | Task package issuer. |
| `signatureRef` | String | Signed package | Signature reference. |
| `signature` | String | Signed package | Signature of the task package. |
| `signatureAlgorithm` | String | Signed package | Package signature algorithm. |
| `signaturePublicKey` | String | Signed package | Public key for verifying the signature. |
| `expiresAt` | String | Signed package | Expiry timestamp. |
| `allowCommit` | Boolean |  | Whether this dispatch permits commits. |
| `allowPush` | Boolean |  | Whether this dispatch permits pushing commits. |
| `allowNetwork` | Boolean |  | Whether network access is permitted. |
| `checkpointRefreshPath` | String | On resume | HTTP path for checkpoint-url. |
| `resumeMode` | String | On resume | Resume mode selected for the dispatch. |
| `resumeSessionBehavior` | String | On resume | FORK branches a session. CANONICAL continues the specified session. |
| `resumeFromDispatchId` | Integer | On resume | Source dispatch for recovery. |
| `resumeProvider` | String | On resume | Provider that owns the native session. |
| `resumeSessionId` | String | On resume | Native session to resume. |
| `resumeCheckpointUrl` | String | On resume | Download URL for the recovery checkpoint. |
| `resumeCheckpointSha256` | String | On resume | SHA-256 digest of the recovery checkpoint. |
| `resumeCheckpointSeq` | Integer | On resume | Sequence number of the recovery checkpoint. |
| `resumeCheckpointCandidates` | Array | On resume | Ordered checkpoint candidates, each with downloadUrl, sha256 and checkpointSeq. |
| `memorySnapshotPath` | String |  | HTTP path for memory/snapshot. |
| `memoryChangesPath` | String |  | HTTP path for memory/changes. |
| `memoryMutationPath` | String |  | HTTP path for memory/mutations. |
| `memoryRecallPath` | String |  | HTTP path for memory/recalls. |
| `memoryStartingRevisions` | Object |  | Starting revision per memory store ID. |
| `modelSelection` | Object | Model selected | Model selection for this dispatch. The model field selects the model ID. |
| `dispatchMcpToken` | String |  | MCP token scoped to the dispatch. |
| `mcpSecrets` | Object | When configured | Per-execution MCP configuration values keyed by opaque references. |
| `environmentVariables` | Object | When configured | Environment variable names and string values for this execution. |
| `debugLog` | Object | Debug logging enabled | Dispatch directive enabling debug logs and setting a byte limit. |

## Example frame

```json
{
  "type": "TASK_DISPATCH",
  "executorId": 42,
  "dispatchId": 701,
  "tenantId": 3,
  "workitemId": 91,
  "agentId": 7,
  "agentVersionId": 11,
  "attempt": 1,
  "downloadUrl": "https://storage.example/task.zip",
  "size": 2048,
  "checksumAlgorithm": "sha256",
  "checksumScope": "zip_archive",
  "checksum": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "artifactUploadPath": "/api/daemon/dispatches/701/artifacts",
  "checkpointUploadPath": "/api/daemon/dispatches/701/checkpoint",
  "packageRefreshPath": "/api/daemon/dispatches/701/package-url"
}
```

## Example reply

See [TASK_ACK](task-ack.md) for the reply fields.

```json
{
  "type": "TASK_ACK",
  "dispatchId": 701
}
```

## Read the task package

Download the ZIP from `downloadUrl`. Check `size`, `checksumAlgorithm`, `checksumScope` and `checksum`; 1.3.0 sends `sha256`, `zip_archive` and a `sha256:`-prefixed digest. The frame also carries `md5`. If the URL expires, POST to `packageRefreshPath` with the executor token. The response contains `downloadUrl` and `expiresInSeconds`. Verify the bytes after downloading from the new URL.

Read `manifest.json` (`schemaVersion: autoWonder.taskPackage.v1`) and its `fileDigests`. Manifest IDs may be decimal strings while the dispatch frame uses JSON integers. Read the task’s actual context: repositories, tools, skills, hooks and policy can vary by package. Signed packages also supply `issuer`, `signatureRef`, `signatureAlgorithm`, `signature`, `signaturePublicKey` and `expiresAt`. The corresponding capability is `TASK_PACKAGE_SIGNATURE_V1`.

## A new dispatch continues earlier work

`resumeFromDispatchId` identifies the source. When `resumeProvider` and `resumeSessionId` identify a local native session, inspect `resumeSessionBehavior`: `FORK` creates a separate child, `CANONICAL` continues the designated session. A valid native session can arrive without any checkpoint URL.

For portable recovery, try `resumeCheckpointCandidates` in order. Each entry contains `downloadUrl`, `sha256` and `checkpointSeq`; the single-checkpoint fields are `resumeCheckpointUrl`, `resumeCheckpointSha256` and `resumeCheckpointSeq`. To renew an expired link, POST to `checkpointRefreshPath` with the executor token and the selected `checkpointSeq`. Verify the sequence and digest. `DEGRADED_CONTINUOUS` can move to another executor without the native session ID.

When the package URL expires, [refresh the package URL](../http/package-url.md). After execution, [upload artifacts](../http/artifacts.md) and [upload the checkpoint](../http/checkpoint.md) before reporting success.

For archive contents and per-file digests see [Task package manifest](../task-package.md).

## Related protocol

- [Accept a dispatch](../tasks/dispatch.md)
- [TASK_ACK](task-ack.md)
- [TASK_BUSY](task-busy.md)
- [TASK_MODEL_SELECTION_V1](../features/task-model-selection-v1.md)
- [AGENT_ENVIRONMENT_VARIABLES_V1](../features/agent-environment-variables-v1.md)
- [DEBUG_LOG_V1](../features/debug-log-v1.md)
- [DISPATCH_REQUESTS_V1](../features/dispatch-requests-v1.md)

Message handling in [TaskDispatchFrame.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/frame/TaskDispatchFrame.java) and [WsDispatchTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchTransport.java).
