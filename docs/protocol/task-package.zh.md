# 任务包清单

[English](task-package.md) | 中文

按派单中的 `downloadUrl` 下载 ZIP，先对完整 ZIP 字节计算 SHA-256，与帧中的 `checksum` 比较，再读取根目录的 `manifest.json`。

## 清单节选

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

示例仅展示与识别和校验有关的字段，摘要为占位值。

| 字段 | 类型 | 描述 |
|---|---|---|
| `schemaVersion` | 字符串 | 此版本固定为 `autoWonder.taskPackage.v1`。 |
| `packageId` | 字符串 | 与派单帧中的包 ID 对应。 |
| `dispatchId` | 字符串 | 十进制派单 ID。外层 WebSocket 帧用 JSON 整数。 |
| `attempt` | 整数 | 派单尝试编号。 |
| `fileDigests` | 对象 | 键为包内路径，值为 `sha256:` 加小写十六进制摘要。按条目解压后的原始字节计算。清单自身不在该映射中。 |

任务正文在 `workitem.md`。仓库列表在 `repos.json` 的 `repos` 字段。执行权限在 `policy.json` 的 `allowCommit`、`allowPush` 和 `allowNetwork` 字段。包内还可能有 `sdlc.json`、`roster.json` 和 `hooks.json`。

生成 ZIP 与逐文件摘要。 [TaskPackager.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/taskpackage/TaskPackager.java)
