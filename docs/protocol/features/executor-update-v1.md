# Remote upgrade

English | [中文](executor-update-v1.zh.md)

Heartbeat capability identifier: `EXECUTOR_UPDATE_V1`.

Supports remote upgrades to a platform-specified target version. Report upgrade phases and the actual version after startup.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "EXECUTOR_UPDATE_V1"
  ]
}
```

## Related protocol

- [EXECUTOR_UPGRADE](../messages/executor-upgrade.md)
- [EXECUTOR_UPGRADE_RESULT](../messages/executor-upgrade-result.md)

Capability implementation in [ExecutorUpdateService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ExecutorUpdateService.java).
