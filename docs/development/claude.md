# Run a Claude Code executor

English | [中文](claude.zh.md)

Add an executor in the maintenance console and select Claude Code under Harness. The executor uses your local Claude Code login and model settings. Importing an AutoWonder command with `--provider claude` also selects Claude Code. The default reported version is `aws-claude-<CLI version>`.

## Start a single executor

Set `AUTOWONDER_TOKEN`, then run:

```sh
go run ./cmd/autowonder-switch -harness claude -list-models
go run ./cmd/autowonder-switch -harness claude -url https://autowonder.example -executor-id 42
```

The daemon looks for `claude` on PATH, then tries `~/.local/bin/claude`. Use `-claude /path/to/claude` to select another installation. The local CLI supplies the model catalog.

## Dispatches and conversations

Dispatches and conversations run in non-interactive mode. The executor provides turn-scoped MCP tools for platform operations and returns the response, session ID and usage. Later turns resume the conversation by session ID. Dispatch packages, result receipts and checkpoints use the executor protocol. Both platform memory and disabled memory turn off Claude's automatic memory.

Read-only tasks can inspect files with Claude Code’s Read, Glob and Grep tools. Writable tasks can also run commands and change files in Claude Code’s built-in Bash sandbox. A writable task fails if the sandbox is unavailable.

The executor does not advertise steering or reliable elicitation answer receipts. If the process exits before a result is saved, the executor cannot confirm completion of that turn from Claude's conversation history.

## Integration tests

The following tests use the installed Claude Code CLI. They start a platform fixture and a model fixture in a temporary environment and check work item handoff, usage, duplicate dispatches, conversation continuity and workspace tool calls.

```sh
AUTOWONDER_TEST_CLAUDE=1 go test -count=1 ./internal/executor -run 'TestDispatchRoundTrip|TestClaudeNativeWorkspace'
```
