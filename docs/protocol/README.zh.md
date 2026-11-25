# AutoWonder 执行器协议

[English](README.md) | 中文

执行器通过 WebSocket 连接 AutoWonder。平台发送派单，执行器确认接单、上报进度并返回结果。任务包通过签名 URL 下载，产物和检查点通过 HTTP 上传。

版本：[社区版 1.3.0](https://github.com/aliyun/alibabacloud-autowonder/releases/tag/1.3.0)

时间：2026 年 11 月 23 日

提交：[0a58555](https://github.com/aliyun/alibabacloud-autowonder/commit/0a585550f1d2369f73b2b8b620c206d8f5ec1e57)

## 跟一张派单走一遍

1. [建立连接](connection.zh.md)。核对认证参数、接入 IP 和关闭码。
2. 发送[心跳](heartbeat.zh.md)，通过[执行清单](execution-inventory.zh.md)上报容量、派单和会话轮次。
3. [接收派单](tasks/dispatch.zh.md)。读取任务包、选择模型并上报进度。
4. [交付结果](tasks/delivery.zh.md)。上传产物和完成检查点，再核对结果回执。
5. [暂停与恢复](tasks/recovery.zh.md)。处理断线、进程重启和会话续接。

## 其他消息

- [会话轮次](conversations.zh.md)按轮次 ID 投递，有独立的回执。
- [能力与控制消息](capabilities.zh.md)说明模型目录、任务指导、交接和远程运维命令。

## 派单 HTTP 接口

- [刷新任务包下载地址](http/package-url.zh.md)
- [上传派单产物](http/artifacts.zh.md)
- [上传检查点](http/checkpoint.zh.md)
- [刷新检查点下载地址](http/checkpoint-url.zh.md)
- [重新认领派单](http/recovery-claim.zh.md)
