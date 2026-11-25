# Task package signatures

English | [中文](task-package-signature-v1.zh.md)

Heartbeat capability identifier: `TASK_PACKAGE_SIGNATURE_V1`.

Supports signed task packages. The dispatch supplies the digest and scope, issuer, signature algorithm, public key, signature and expiry. See the dispatch guide for the fields.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "TASK_PACKAGE_SIGNATURE_V1"
  ]
}
```

## Signed bytes

An executor advertising this feature can still receive an unsigned package. Verify the archive against `checksum`. If a signature is present, verify it as described below.

For a signed package, encode the six lines below as UTF-8. Terminate every line, including the last, with LF. Substitute the original frame values for the placeholders and use the lowercase hex SHA-256 of the archive on the final line.

```text
autowonder.taskPackage.signature.v1
{issuer}
{signatureRef}
{expiresAt}
{packageId}
sha256:{archiveSha256Hex}
```

The algorithm identifier is `ed25519`. Decode `signaturePublicKey` from Base64 to DER SubjectPublicKeyInfo. The `signatureRef` value is `sha256:` plus the hex SHA-256 of that DER. Decode `signature` from Base64 and verify it over the bytes above. [TaskPackageSigner.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/taskpackage/TaskPackageSigner.java) · [ObjectStorageConfig.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/storage/ObjectStorageConfig.java) · [TaskPackager.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/taskpackage/TaskPackager.java)

## Related protocol

- [Accept a dispatch](../tasks/dispatch.md)

Capability implementation in [WsDispatchTransport.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/websocket/WsDispatchTransport.java).
