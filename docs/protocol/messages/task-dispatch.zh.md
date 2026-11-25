# 下发派单

[English](task-dispatch.md) | 中文

平台 → 执行器。帧类型为 `TASK_DISPATCH`。

平台用 `dispatchId` 标识派单、用 `executorId` 指定接收者。能接单时先回 `TASK_ACK`，再从 `downloadUrl` 下载任务包并校验摘要。无法接单时回 `TASK_BUSY`。

## 消息参数

| 字段 | JSON 类型 | 适用条件 | 描述 |
| --- | --- | --- | --- |
| `type` | 字符串 |  | 固定为 `TASK_DISPATCH`。 |
| `executorId` | 整数 |  | 该连接对应的执行器 ID。 |
| `dispatchId` | 整数 |  | 本次派单 ID，同一工单重新派发时可能不同。 |
| `tenantId` | 整数 |  | 工单所属租户的 ID。 |
| `workitemId` | 整数 |  | 派单所属工单的 ID。 |
| `agentId` | 整数 |  | 执行本次任务的智能体 ID。 |
| `agentVersionId` | 整数 |  | 本次执行使用的智能体版本 ID。 |
| `attempt` | 整数 |  | 本次派单的尝试次数。 |
| `downloadUrl` | 字符串 |  | 下载文件的签名 URL。 |
| `size` | 整数 |  | 文件大小，单位为字节。 |
| `checksumAlgorithm` | 字符串 |  | 包摘要使用的算法。 |
| `checksumScope` | 字符串 |  | 摘要覆盖的内容，此值表示整个 ZIP。 |
| `checksum` | 字符串 |  | 任务包的 SHA-256 摘要。 |
| `artifactUploadPath` | 字符串 |  | 上传产物的 HTTP 路径。 |
| `checkpointUploadPath` | 字符串 |  | 上传检查点的 HTTP 路径。 |
| `packageRefreshPath` | 字符串 |  | 刷新任务包下载地址的 HTTP 路径。 |
| `sdlcStepId` | 整数 |  | 本次派单对应的流程步骤 ID。 |
| `idempotencyKey` | 字符串 |  | 重发同一操作时沿用的幂等键。 |
| `md5` | 字符串 |  | 任务包的 MD5 十六进制摘要。 |
| `packageId` | 字符串 |  | 任务包标识。 |
| `issuer` | 字符串 | 任务包带签名时 | 任务包签发方。 |
| `signatureRef` | 字符串 | 任务包带签名时 | 签名引用。 |
| `signature` | 字符串 | 任务包带签名时 | 任务包签名。 |
| `signatureAlgorithm` | 字符串 | 任务包带签名时 | 签名算法标识。 |
| `signaturePublicKey` | 字符串 | 任务包带签名时 | 验签使用的公钥。 |
| `expiresAt` | 字符串 | 任务包带签名时 | 有效期截止时间。 |
| `allowCommit` | 布尔值 |  | 本次任务是否允许提交代码。 |
| `allowPush` | 布尔值 |  | 本次任务是否允许推送代码。 |
| `allowNetwork` | 布尔值 |  | 本次任务是否允许访问网络。 |
| `checkpointRefreshPath` | 字符串 | 恢复已有任务时 | 刷新检查点下载地址的 HTTP 路径。 |
| `resumeMode` | 字符串 | 恢复已有任务时 | 本次派单采用的恢复模式。 |
| `resumeSessionBehavior` | 字符串 | 恢复已有任务时 | FORK 分出子会话，CANONICAL 续接指定会话。 |
| `resumeFromDispatchId` | 整数 | 恢复已有任务时 | 恢复来源派单的 ID。 |
| `resumeProvider` | 字符串 | 恢复已有任务时 | 持有原生会话的 provider。 |
| `resumeSessionId` | 字符串 | 恢复已有任务时 | 需要续接的原生会话 ID。 |
| `resumeCheckpointUrl` | 字符串 | 恢复已有任务时 | 恢复检查点的下载地址。 |
| `resumeCheckpointSha256` | 字符串 | 恢复已有任务时 | 恢复检查点的 SHA-256 摘要。 |
| `resumeCheckpointSeq` | 整数 | 恢复已有任务时 | 恢复检查点的序号。 |
| `resumeCheckpointCandidates` | 数组 | 恢复已有任务时 | 按尝试顺序排列的检查点，每项含 downloadUrl、sha256 和 checkpointSeq。 |
| `memorySnapshotPath` | 字符串 |  | 读取记忆快照的 HTTP 路径。 |
| `memoryChangesPath` | 字符串 |  | 读取记忆变更的 HTTP 路径。 |
| `memoryMutationPath` | 字符串 |  | 写入记忆变更的 HTTP 路径。 |
| `memoryRecallPath` | 字符串 |  | 上报记忆召回的 HTTP 路径。 |
| `memoryStartingRevisions` | 对象 |  | 按记忆库 ID 记录的起始修订号。 |
| `modelSelection` | 对象 | 派单指定模型时 | 本次派单的模型选择，model 指定模型 ID。 |
| `dispatchMcpToken` | 字符串 |  | 该派单访问 MCP 服务的令牌。 |
| `mcpSecrets` | 对象 | 随本次配置 | 以不透明引用为键的本次 MCP 配置值。 |
| `environmentVariables` | 对象 | 随本次配置 | 本次执行的环境变量名与字符串值。 |
| `debugLog` | 对象 | 调试日志已开启 | 派单时指定是否采集调试日志及字节上限。 |

## 消息示例

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

## 对应回包

回包参数见 [TASK_ACK](task-ack.zh.md)。

```json
{
  "type": "TASK_ACK",
  "dispatchId": 701
}
```

## 读取任务包

从 `downloadUrl` 下载 ZIP。校验 `size`、`checksumAlgorithm`、`checksumScope` 和 `checksum`；1.3.0 使用 `sha256` 算法，对 `zip_archive` 求摘要，摘要带 `sha256:` 前缀。帧也保留 `md5`。链接过期后，带执行器令牌向 `packageRefreshPath` 发 POST；从响应取 `downloadUrl` 和 `expiresInSeconds`，重新下载后仍须校验。

读包内 `manifest.json`（`schemaVersion: autoWonder.taskPackage.v1`）和逐文件摘要 `fileDigests`。清单里的 ID 可以是十进制字符串，外层帧里的 ID 是 JSON 整数。按本次包内的仓库、工具、技能、hook 和策略准备环境。签名包还会给出 `issuer`、`signatureRef`、`signatureAlgorithm`、`signature`、`signaturePublicKey`、`expiresAt`；对应能力为 `TASK_PACKAGE_SIGNATURE_V1`。

## 新派单续接已有工作

`resumeFromDispatchId` 指明来源。`resumeProvider` 和 `resumeSessionId` 指向本机原生会话时，还要看 `resumeSessionBehavior`：`FORK` 从父会话分出独立子会话，`CANONICAL` 续接指定会话。平台可以只给原生会话，不给检查点 URL。

可移植恢复按顺序尝试 `resumeCheckpointCandidates`，每项含 `downloadUrl`、`sha256`、`checkpointSeq`。单检查点字段为 `resumeCheckpointUrl`、`resumeCheckpointSha256`、`resumeCheckpointSeq`。链接过期时，用执行器令牌和选定的 `checkpointSeq` 向 `checkpointRefreshPath` 发 POST，核对返回的序号和摘要。`DEGRADED_CONTINUOUS` 可能换机执行，不携带原生会话 ID。

任务包 URL 过期时调用[刷新任务包下载地址](../http/package-url.zh.md)。执行完成后通过[上传产物](../http/artifacts.zh.md)和[上传检查点](../http/checkpoint.zh.md)准备结果。

任务包内的文件和摘要格式见[任务包清单](../task-package.zh.md)。

## 相关协议

- [接收派单](../tasks/dispatch.zh.md)
- [TASK_ACK](task-ack.zh.md)
- [TASK_BUSY](task-busy.zh.md)
- [TASK_MODEL_SELECTION_V1](../features/task-model-selection-v1.zh.md)
- [AGENT_ENVIRONMENT_VARIABLES_V1](../features/agent-environment-variables-v1.zh.md)
- [DEBUG_LOG_V1](../features/debug-log-v1.zh.md)
- [DISPATCH_REQUESTS_V1](../features/dispatch-requests-v1.zh.md)

消息处理见[TaskDispatchFrame.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/frame/TaskDispatchFrame.java)、[WsDispatchTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchTransport.java)。
