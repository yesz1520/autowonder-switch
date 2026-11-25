# 暂停、取消与恢复

[English](recovery.md) | 中文

收到暂停命令时保存恢复检查点。执行器重连后可重新认领派单，后续派单也可通过检查点继续执行。

## 接口与事件

- [暂停或取消派单](../messages/task-pause.zh.md)
- [报告已暂停](../messages/task-paused.zh.md)
- [报告暂停失败](../messages/task-pause-failed.zh.md)
- [重新认领派单](../http/recovery-claim.zh.md)
- [刷新检查点下载地址](../http/checkpoint-url.zh.md)
- [TASK_DISPATCH](../messages/task-dispatch.zh.md)
