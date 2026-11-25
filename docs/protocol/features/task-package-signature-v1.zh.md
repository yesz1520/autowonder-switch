# 任务包验签

[English](task-package-signature-v1.md) | 中文

心跳能力标识为 `TASK_PACKAGE_SIGNATURE_V1`。

支持校验签名任务包。派单携带摘要及作用范围、签发方、签名算法、公钥、签名和过期时间，字段分别见接收派单。

## 声明示例

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "TASK_PACKAGE_SIGNATURE_V1"
  ]
}
```

## 签名输入

声明此能力后仍可能收到未签名的任务包。用 `checksum` 校验整包；收到签名时，再按下文验签。

收到签名包时，将下面六行编码为 UTF-8。每行以 LF 结束，包括最后一行。花括号表示替换为帧里的原值，最后一行使用整包 SHA-256 的小写十六进制值。

```text
autowonder.taskPackage.signature.v1
{issuer}
{signatureRef}
{expiresAt}
{packageId}
sha256:{archiveSha256Hex}
```

算法标识为 `ed25519`。`signaturePublicKey` 是 Base64 编码的 DER 公钥（SubjectPublicKeyInfo），`signatureRef` 是 `sha256:` 加该 DER 的 SHA-256 十六进制摘要。将 `signature` 从 Base64 解码后，用公钥验证上述字节。 [TaskPackageSigner.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/taskpackage/TaskPackageSigner.java) · [ObjectStorageConfig.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/storage/ObjectStorageConfig.java) · [TaskPackager.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/taskpackage/TaskPackager.java)

## 相关协议

- [接收派单](../tasks/dispatch.zh.md)

能力定义与使用见[WsDispatchTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchTransport.java)。
