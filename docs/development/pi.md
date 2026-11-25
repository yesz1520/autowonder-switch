# Run a pi executor

English | [中文](pi.zh.md)

Add an executor in the console and select pi under Harness. Importing a connect command with `--provider pi` also selects it. The reported version defaults to `aws-pi-<CLI version>`. The daemon reads the model catalog and default selection from the local pi RPC interface. Use `provider/model` to distinguish models with the same ID across providers.

## Startup

Configure a model and its credentials in pi first. Single-executor mode reads the platform token from `AUTOWONDER_TOKEN`:

```sh
go run ./cmd/autowonder-switch -harness pi -list-models
go run ./cmd/autowonder-switch -harness pi -url https://autowonder.example -executor-id 42
```

The daemon looks for `pi` on PATH. Select another executable with `-pi /path/to/pi`. Native dispatch and conversation sessions live under the executor's state directory. Later turns resume by session ID and forks receive a new ID.

## Tools and memory

Each turn gets a local tool service for platform calls and user questions. Installed extensions, MCP servers and skills are excluded from discovery. With provider-local memory, pi loads its context files as configured. Platform memory and disabled memory both turn off context file discovery.

Workspace commands run under Seatbelt on macOS and bubblewrap on Linux (amd64 or arm64). On Linux, install bubblewrap 0.8 or later at `/usr/bin/bwrap`; use a non-setuid build on a host that permits unprivileged user namespaces. If the sandbox cannot start, the command fails. Text conversations only need pi.

Commands can read the workspace, system tools, and the repository metadata and packaged skills provided for the task. Writable tasks can also change workspace files. Each command gets a temporary directory for scratch files, with network access blocked. The daemon passes task environment variables into the sandbox. Tools installed under the user's home directory may be inaccessible.

Package tool hooks, native plugins and reliable elicitation answer receipts are unsupported. A saved session can be resumed. The adapter cannot recover an unsaved turn result after process exit.

## Integration tests

The tests use the installed pi CLI, a temporary model configuration and a local platform fixture. They cover work item handoff, usage, result replay, conversation continuity, workspace tools and session forks.

```sh
AUTOWONDER_TEST_PI=1 go test -count=1 ./internal/executor -run 'TestDispatchRoundTrip|TestPiNative'
```
