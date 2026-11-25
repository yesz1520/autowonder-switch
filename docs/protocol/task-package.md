# Task package manifest

English | [中文](task-package.zh.md)

Download the ZIP from the dispatch’s `downloadUrl`. Compute SHA-256 over the entire ZIP and compare it with the frame’s `checksum`, then read `manifest.json` at the archive root.

## Manifest excerpt

```json
{
  "schemaVersion": "autoWonder.taskPackage.v1",
  "packageId": "pkg_701",
  "dispatchId": "701",
  "attempt": 1,
  "fileDigests": {
    "workitem.md": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
  }
}
```

The excerpt shows the identification and verification fields. The digest is illustrative.

| Field | Type | Description |
|---|---|---|
| `schemaVersion` | String | Fixed at `autoWonder.taskPackage.v1` for this version. |
| `packageId` | String | Matches the package ID in the dispatch frame. |
| `dispatchId` | String | Decimal dispatch ID. The outer WebSocket frame uses a JSON integer. |
| `attempt` | Integer | Dispatch attempt number. |
| `fileDigests` | Object | Maps archive paths to `sha256:` plus a lowercase hex digest of the uncompressed entry bytes. The manifest itself is excluded. |

Read the task from `workitem.md`, repository definitions from `repos.json` under `repos`, and permissions from `allowCommit`, `allowPush` and `allowNetwork` in `policy.json`. The archive can also include `sdlc.json`, `roster.json` and `hooks.json`.

Builds the ZIP and its per-file digest map. [TaskPackager.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/taskpackage/TaskPackager.java)
