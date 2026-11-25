# 上传检查点

[English](checkpoint.md) | 中文

`POST /api/daemon/dispatches/{dispatchId}/checkpoint`

通过 checkpointUploadPath 保存归档。下例上传完成检查点，随后用相同的 resultId 上报任务结果。

## 请求参数

| 参数 | 位置 | 类型 | 必填 | 描述 |
| --- | --- | --- | --- | --- |
| `dispatchId` | path | 整数 | 是 | 本次派单 ID，同一工单重新派发时可能不同。 |
| `token` | query | 字符串 | 是 | 该派单执行器的连接令牌。 |
| `checkpointSeq` | multipart | 整数 | 是 | 该派单的检查点序号，使用正整数。 |
| `checkpoint` | multipart | File | 是 | 非空的检查点归档文件。 |
| `purpose` | multipart | 字符串 | 可选 | 完成任务用 COMPLETION；保存恢复点用 RECOVERY 或省略。 |
| `resultId` | multipart | 字符串 | 按条件 | purpose 为 COMPLETION 时必填；恢复检查点省略。 |
| `provider` | multipart | 字符串 | 可选 | 生成检查点的 provider。 |
| `providerSessionId` | multipart | 字符串 | 可选 | 与检查点关联的原生会话 ID。 |
| `runtimeId` | multipart | 字符串 | 可选 | 生成检查点的运行时标识。 |
| `activeStepId` | multipart | 字符串 | 可选 | 保存检查点时的活动步骤。 |

## 请求示例

```sh
curl -X POST 'https://platform.example/api/daemon/dispatches/701/checkpoint?token=EXAMPLE' -F 'checkpointSeq=2' -F 'checkpoint=@checkpoint.zip' -F 'purpose=COMPLETION' -F 'resultId=86ebc809-2d32-44e4-b456-29b870f878bd'
```

## 响应参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `checkpointSeq` | 整数 | 该派单的检查点序号，使用正整数。 |
| `sha256` | 字符串 | 检查点归档的 SHA-256 摘要。 |
| `sizeBytes` | 整数 | 上传文件的大小，单位为字节。 |
| `purpose` | 字符串 | 完成任务用 COMPLETION；保存恢复点用 RECOVERY 或省略。 仅完成检查点返回。 |
| `resultId` | 字符串 | 本次结果的小写 UUID；重复交付沿用原值。 仅完成检查点返回。 |

## 成功响应

```json
{
  "checkpointSeq": 2,
  "sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "sizeBytes": 2048,
  "purpose": "COMPLETION",
  "resultId": "86ebc809-2d32-44e4-b456-29b870f878bd"
}
```

检查点文件最大为 50 MiB。检查点为空、超过大小限制或序号非正数时返回 HTTP 400。

普通表单参数也可放在查询字符串中，文件仍用 multipart 的 `checkpoint` 字段传送。相同派单、相同 `checkpointSeq` 重传时必须保持字节一致，否则返回 HTTP 409。需要随检查点上报元数据时，将 `checkpoint.json` 放在 tar.gz 归档根目录。 [DispatchCheckpointService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DispatchCheckpointService.java)

## 相关源码

处理请求并生成响应。 [DaemonCheckpointController.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DaemonCheckpointController.java)
