# 上传派单产物

[English](artifacts.md) | 中文

`POST /api/daemon/dispatches/{dispatchId}/artifacts`

用 artifactUploadPath 上传一个或多个文件。逐项检查返回的 files，HTTP 200 也可能包含被拒绝的文件。

## 请求参数

| 参数 | 位置 | 类型 | 必填 | 描述 |
| --- | --- | --- | --- | --- |
| `dispatchId` | path | 整数 | 是 | 本次派单 ID，同一工单重新派发时可能不同。 |
| `token` | query | 字符串 | 是 | 该派单执行器的连接令牌。 |
| `files` | multipart | File | 是 | multipart 文件字段，可重复传多个文件。 |
| `filesMetadata` | multipart | 字符串 | 可选 | 与 files 对应的 JSON 数组字符串，可附文件路径等元数据。 |
| `idempotencyKey` | multipart | 字符串 | 可选 | 重发同一操作时沿用的幂等键。 |

## 请求示例

```sh
curl -X POST 'https://platform.example/api/daemon/dispatches/701/artifacts?token=EXAMPLE' -F 'files=@review.md'
```

## 响应参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `remoteRef` | 字符串 | 本次上传的存储前缀。 |
| `files` | 数组 | 逐文件回执。index 从 0 开始，path 为文件路径，status 为 ACCEPTED 或 REJECTED。接受时返回存储引用和 sizeBytes，拒绝时返回 code，可能附 maxBytes。 |

## 成功响应

```json
{
  "remoteRef": "example/701/",
  "files": [
    {
      "index": 0,
      "path": "review.md",
      "status": "ACCEPTED",
      "sizeBytes": 2048,
      "ossRef": "example/701/review.md",
      "remoteRef": "example/701/review.md"
    }
  ]
}
```

一次请求最多上传 200 个文件。超出数量限制时返回 HTTP 400。

每个文件最大 50 MiB。`filesMetadata` 用 multipart 文本字段发送，值是 JSON 数组字符串。按上传顺序用数组下标关联 `files`；未提供对应的 `path` 时，平台使用 multipart 文件名。下例中第二个文件被拒绝，只需从该项的 `code` 读取原因。

```json
{
  "files": [
    {
      "index": 0,
      "path": "review.md",
      "status": "ACCEPTED",
      "sizeBytes": 200,
      "ossRef": "example/review.md",
      "remoteRef": "example/review.md"
    },
    {
      "index": 1,
      "path": "large.bin",
      "status": "REJECTED",
      "code": "FILE_TOO_LARGE",
      "sizeBytes": 52428801,
      "maxBytes": 52428800
    }
  ]
}
```

## 相关源码

处理请求并生成响应。 [DaemonArtifactController.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/artifact/DaemonArtifactController.java)
