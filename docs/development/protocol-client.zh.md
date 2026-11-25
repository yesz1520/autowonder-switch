# 接入 Go 协议包

[English](protocol-client.md) | 中文

仓库内的 daemon 通过 `internal/autowonder` 与平台通信。一个 `Client` 对应一个平台上的一个执行器 ID。

## 建立连接

用 `New(Config{BaseURL, ExecutorID, Token, ...})` 创建客户端。`BaseURL` 填平台的 HTTP(S) 基地址。为这个客户端调用一次 `OpenOutbox`，再把它和心跳快照函数传给 `NewEndpoint`。不同身份可共用一个状态目录，文件按平台和执行器 ID 隔离。

用 `On` 注册要处理的下行消息，在心跳中声明能实际处理的能力，然后调用 `client.Run`。连接后立即发送心跳，网络中断会重连。鉴权失败或同一身份的连接被替换时，`Run` 返回错误。模型目录的完整接线见[Go 示例](../../internal/autowonder/example_test.go)。

## 接单与结果交付

收到 `TaskDispatch` 时，先判断本机是否能接单。接单后回 `TaskAck`，交给任务工作线程下载和校验包，再执行。回调应尽快返回，以便连接继续接收暂停、指导和回执。任务的 context 由执行器管理；连接回调的 context 在断线时结束。

V2 结果先生成并保存 `NewResultID()`。成功时用该 ID 上传完成检查点，再把返回的序号和摘要填入 `TaskResult`。用 `endpoint.Send` 交付结果：先保存原帧，再写 WebSocket。离线时传入 nil session 可先排队。回执匹配后自动清理，`RETRYABLE` 会保留；需要在当前连接重试时调用 `outbox.Replay`。

## 调用 HTTP 接口

任务包下载、检查点、产物和诊断上传、任务数据、人工请求、用量和环境快照都通过同一 `Client` 调用。记忆接口使用 `client.Memory(dispatchID, dispatchMCPToken)`，传入本次派单的凭据。HTTP 失败可用 `errors.As` 读取 `*HTTPError` 的状态码和响应体。上传产物时即使 HTTP 成功，也要检查逐文件回执。

消息方向、字段和样例见[执行器协议](../protocol/README.zh.md)。
