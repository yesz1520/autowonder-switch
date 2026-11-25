# 运行 Claude Code 执行器

[English](claude.md) | 中文

在维护页添加执行器，Harness 选「Claude Code」。执行器沿用本机 Claude Code 的登录和模型配置。导入带有 `--provider claude` 的 AutoWonder 启动指令也会选中它。默认上报版本为 `aws-claude-<CLI 版本>`。

## 单执行器启动

先设置 `AUTOWONDER_TOKEN`，再运行：

```sh
go run ./cmd/autowonder-switch -harness claude -list-models
go run ./cmd/autowonder-switch -harness claude -url https://autowonder.example -executor-id 42
```

程序先从 PATH 查找 `claude`，然后尝试 `~/.local/bin/claude`。用 `-claude /path/to/claude` 可以指定其他安装位置。模型列表从本机 CLI 读取。

## 工单与会话

工单和会话以非交互模式运行。执行器通过当前轮次的 MCP 工具访问平台，并将正文、会话 ID 和用量交回平台。后续轮次通过会话 ID 恢复对话。工单包、结果回执和检查点沿用执行器协议。平台记忆和关闭记忆模式都会关闭 Claude 的自动记忆。

工作区命令依赖本机的 Codex 沙箱工具。普通文本对话只需要 Claude Code。当前未声明运行中追加指令和可靠追问答复回执能力。进程退出前如果尚未保存结果，执行器不能从 Claude 的历史会话确认该轮已完成。

## 集成测试

以下测试使用已安装的 Claude Code CLI。测试在临时目录启动模拟平台和模型服务，检查工单交接、用量、重复派单、会话续聊以及工作区工具调用。

```sh
AUTOWONDER_TEST_CLAUDE=1 go test -count=1 ./internal/executor -run 'TestDispatchRoundTrip|TestClaudeNativeWorkspace'
```
