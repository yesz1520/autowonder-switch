# 刷新任务包下载地址

[English](package-url.md) | 中文

`POST /api/daemon/dispatches/{dispatchId}/package-url`

任务包链接过期时，用派单的 packageRefreshPath 重新获取下载地址。

## 请求参数

| 参数 | 位置 | 类型 | 必填 | 描述 |
| --- | --- | --- | --- | --- |
| `dispatchId` | path | 整数 | 是 | 本次派单 ID，同一工单重新派发时可能不同。 |
| `token` | query | 字符串 | 是 | 该派单执行器的连接令牌。 |

## 请求示例

```sh
curl -X POST 'https://platform.example/api/daemon/dispatches/701/package-url?token=EXAMPLE'
```

## 响应参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `downloadUrl` | 字符串 | 下载文件的签名 URL。 |
| `expiresInSeconds` | 整数 | 新签名 URL 的剩余有效秒数。 |

## 成功响应

```json
{
  "downloadUrl": "https://storage.example/task.zip",
  "expiresInSeconds": 900
}
```

派单的任务包无法续签时返回 HTTP 409。

## 相关源码

处理请求并生成响应。 [DaemonTaskPackageController.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/taskpackage/DaemonTaskPackageController.java)
