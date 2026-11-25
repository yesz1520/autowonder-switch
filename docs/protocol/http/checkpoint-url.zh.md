# 刷新检查点下载地址

[English](checkpoint-url.md) | 中文

`POST /api/daemon/dispatches/{dispatchId}/checkpoint-url`

用派单的 checkpointRefreshPath 和选中的检查点序号刷新下载链接。

## 请求参数

| 参数 | 位置 | 类型 | 必填 | 描述 |
| --- | --- | --- | --- | --- |
| `dispatchId` | path | 整数 | 是 | 本次派单 ID，同一工单重新派发时可能不同。 |
| `token` | query | 字符串 | 是 | 该派单执行器的连接令牌。 |
| `checkpointSeq` | query | 整数 | 是 | 该派单的检查点序号，使用正整数。 |

## 请求示例

```sh
curl -X POST 'https://platform.example/api/daemon/dispatches/701/checkpoint-url?token=EXAMPLE&checkpointSeq=2'
```

## 响应参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `checkpointSeq` | 整数 | 该派单的检查点序号，使用正整数。 |
| `downloadUrl` | 字符串 | 下载文件的签名 URL。 |
| `expiresAt` | 字符串 | 有效期截止时间。 |
| `expiresInSeconds` | 整数 | 新签名 URL 的剩余有效秒数。 |
| `sha256` | 字符串 | 检查点归档的 SHA-256 摘要。 |

## 成功响应

```json
{
  "checkpointSeq": 2,
  "downloadUrl": "https://storage.example/checkpoint.zip",
  "expiresAt": "2026-11-25T09:00:00Z",
  "expiresInSeconds": 900,
  "sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
```

## 相关源码

处理请求并生成响应。 [DaemonCheckpointController.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DaemonCheckpointController.java)
