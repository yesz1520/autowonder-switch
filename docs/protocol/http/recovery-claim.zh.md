# 重新认领派单

[English](recovery-claim.md) | 中文

`POST /api/daemon/dispatches/{dispatchId}/recovery-claim`

执行器恢复后，用原 dispatchId 认领仍在执行的派单。可认领状态为 DISPATCHED、ACKED 和 RUNNING。

## 请求参数

| 参数 | 位置 | 类型 | 必填 | 描述 |
| --- | --- | --- | --- | --- |
| `dispatchId` | path | 整数 | 是 | 本次派单 ID，同一工单重新派发时可能不同。 |
| `token` | query | 字符串 | 是 | 该派单执行器的连接令牌。 |

## 请求示例

```sh
curl -X POST 'https://platform.example/api/daemon/dispatches/701/recovery-claim?token=EXAMPLE'
```

## 响应参数

| 字段 | JSON 类型 | 描述 |
| --- | --- | --- |
| `allowed` | 布尔值 | 平台允许当前执行器继续执行该派单。 |
| `status` | 字符串 | 派单当前状态。 |
| `environmentVariables` | 对象 | 本次执行的环境变量名与字符串值。 |

## 成功响应

```json
{
  "allowed": true,
  "status": "RUNNING",
  "environmentVariables": {
    "LANG": "C.UTF-8"
  }
}
```

HTTP 404 表示派单不存在，409 表示当前状态不允许认领，401 表示认证或写入权限校验失败。

## 相关源码

处理请求并生成响应。 [DaemonRecoveryClaimController.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/dispatch/DaemonRecoveryClaimController.java)
