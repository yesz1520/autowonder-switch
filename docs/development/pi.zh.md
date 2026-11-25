# 运行 pi 执行器

[English](pi.md) | 中文

在维护页添加执行器，Harness 选「pi」。导入带 `--provider pi` 的连接指令也会选中它。版本默认上报为 `aws-pi-<CLI 版本>`，模型目录与默认模型从本机 pi 的 RPC 接口读取。平台下发模型时，可用 `provider/model` 区分不同服务商的同名模型。

## 启动

先在本机配置好 pi 的模型和凭据。单执行器模式从环境变量 `AUTOWONDER_TOKEN` 读取平台令牌：

```sh
go run ./cmd/autowonder-switch -harness pi -list-models
go run ./cmd/autowonder-switch -harness pi -url https://autowonder.example -executor-id 42
```

默认从 PATH 查找 `pi`。用 `-pi /path/to/pi` 可以指定可执行文件。工单和会话的原生记录分别归入对应执行器的状态目录，续聊使用原会话 ID，分叉产生新的 ID。

## 工具和记忆

平台工具和追问通过每轮独立的本机工具服务调用。安装的扩展、MCP 和技能不会自动加载。选择「本机 Agent 记忆」时，pi 按本机规则加载上下文文件；平台记忆和关闭记忆模式均关闭上下文文件发现。

macOS 使用 Seatbelt 执行工作区命令，Linux（amd64、arm64）使用 bubblewrap。Linux 主机需要允许非特权用户命名空间，并在 `/usr/bin/bwrap` 安装未设置 setuid 位的 0.8 或更高版本。沙箱启动失败时，命令会报错。纯文本对话只需要 pi。

命令可以读取工作区、系统工具，以及本次任务提供的仓库元数据和技能包。可写任务还可以修改工作区文件。每条命令有独立的临时目录，网络访问被禁止。daemon 会将任务环境变量传入沙箱。安装在用户主目录下的工具可能无法访问。

暂不支持任务包的工具前后置 hook、原生插件或可靠追问答复回执。原生会话可续聊，但进程退出前未保存的轮次结果不能从当前适配器自动找回。

## 集成测试

测试使用真实 pi CLI、临时模型配置和本地模拟平台，覆盖工单交接、用量、结果补发、会话续聊、工作区工具和会话分叉。

```sh
AUTOWONDER_TEST_PI=1 go test -count=1 ./internal/executor -run 'TestDispatchRoundTrip|TestPiNative'
```
