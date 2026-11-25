# Run a Codex executor

English | [中文](codex.zh.md)

The executor uses your local Codex configuration and login. List the available models first. This starts the Codex app-server without connecting to AutoWonder or running a model turn.

```sh
go run ./cmd/autowonder-switch -list-models
```

## Manage multiple executors on one host

Start the daemon and open `http://127.0.0.1:8787`. Choose Add Executor and enter a name, platform URL, executor ID, token and capacity. Each identity has its own connection and uses the local Codex configuration. Stopping one interrupts its active tasks while the other identities keep running.

```sh
go run ./cmd/autowonder-switch
```

Use `-listen` to change the loopback port. The daemon saves the configuration and tokens in `executors.json` in the state directory with owner-only permissions. Enabled identities reconnect on startup. To rotate a token, stop the identity, edit and save the connection, then enable it again. A platform restart request restarts only its target identity.

The Go binary embeds the Vue 3 maintenance page. After editing the frontend, rebuild the assets from the repository root before building the Go binary.

```sh
npm ci --prefix internal/daemon/ui
npm run build --prefix internal/daemon/ui
```

## Connect an executor

Set the executor token in `AUTOWONDER_TOKEN`. Substitute your platform URL and executor ID below. The process runs in the foreground until you press Ctrl+C. Capacity defaults to 1 and accepts up to 50 concurrent dispatches and conversation turns combined.

```sh
go run ./cmd/autowonder-switch -url https://autowonder.example -executor-id 42 -capacity 2
```

Supplying the platform URL and executor ID starts the single-executor mode. State defaults to `autowonder-switch` under the user configuration directory, partitioned by platform and identity. Use `-state-dir` to choose another directory and `-codex` to select an executable. Active tasks survive WebSocket reconnects. Exiting the process interrupts model turns. A repeated dispatch can receive its saved final reply.

When the platform sends an upgrade target, Switch saves that version for the current executor identity and reports it in subsequent heartbeats, including after a restart. Operators update the local Codex binary separately.

On Linux and macOS, a remote restart starts the executor again with its existing arguments and environment. With `WHEN_IDLE` it stops accepting work and waits for current dispatches and conversation turns to finish. The first heartbeat after restart includes the request ID.

AutoWonder can apply launch settings to an executor on Linux or macOS. After restart, the handshake and first heartbeat use the new capacity. The heartbeat includes the request ID and target runtime version. New tasks and conversations get separate directories under `workspaceRoot`. Existing workspaces keep their paths. The executor uses `daemonName` in its logs and logs incoming frame types when `debug` is enabled.

With `memoryMode` set to `platform`, the executor provides platform memory. With `provider-local`, Codex uses its local memory settings. With `none`, the executor disables memory injection. AutoWonder 1.3.0 clears the model, reasoning effort and context window when preparing launch settings for `CODEX_CLI`. An individual dispatch can still select a model with `modelSelection.model`.

## Run a work item

For text tasks, the executor reads the verified package, uploads the answer as a `result.md` artifact, then submits a completion checkpoint and a V2 result. A dispatch can select a model, steer the running turn or cancel it. Text tasks support pause, resume or fork from a Codex session owned by this executor, and recovery from a text checkpoint written by this executor.

When Codex reports token usage, the executor sends it by dispatch, workflow step and model. If a reviewer requests changes, both turns count toward the same step. A conversation reports usage for the current turn in its receipt.

When a work item calls for a handoff, Codex can use `executor.request_handoff` to choose an agent or human. Switch includes the proposal in the successful task result. The platform selects the next owner. Its decision is saved as `handoff-result.json` in the dispatch directory.

If the executor exits before sending the result and the platform resends the same dispatch attempt, Switch can finish delivery from a saved workflow reply. For a single turn with no pending approval or exit hook, it can also read the completed answer from the local Codex history using both the session ID and turn ID.

When a dispatch supplies memory routes and a dispatch token, Codex can list stores and document versions, read a document, and fetch changes after a store revision. Reading a document records a recall when the dispatch provides that route. Writes carry an idempotency key. Updates and deletions also carry the version Codex read so the platform can reject conflicting edits.

Before the first heartbeat after a restart, Switch recovers final task results and V1 conversation replies interrupted between saving the answer and queuing it for acknowledgement. The platform can receive the saved answer without resending the original work.

When a workflow step enables `askWhenUnclear`, Codex can open a form on the work item and continue the same turn with the answer. With `humanApprovalRequired`, the executor submits the step for review, including a SHA-256 fingerprint of the answer and repository snapshot. A revision request goes back to Codex while the step has retries left. The executor waits for approval before advancing or delivering the result.

When a dispatch enables `debugLog.enabled`, the executor records stage timestamps, the dispatch ID and message byte counts within `maxBytes`. It uploads the compressed log to the signed URL and falls back to the platform relay. The task result includes the upload outcome.

Task packages can run hooks around repository preparation, the Codex run and workflow steps, before a commit or push, and on failure or cleanup. Switch verifies each hook bundle digest, then runs its scripts in the Codex sandbox according to `failurePolicy` and `onceScope`. Output is saved under `hooks-<attempt>/logs/<trigger>/<name>/` in the dispatch directory.

Switch maps `beforeTool` and `afterTool` to the native Codex events `PreToolUse` and `PostToolUse`. It advertises tool hooks after a successful startup probe. A pre-hook can deny the pending call. By the time a post-hook receives the result, the tool has already run.

### Code work items

The executor reads `repos.json`. When network access is allowed, it checks out the requested revision from the HTTP(S) or SSH URL. Private repositories use the daemon user's configured Git credentials or SSH agent. An existing local copy can be mapped as shown below. Offline checkout requires the revision to exist locally. An unavailable `pinnedBaseCommit` fails the dispatch. Each dispatch has a separate working directory where Codex can inspect Git status, diffs and history, and run project tests.

Codex selects patch, local commit or push for each repository from the work item. The executor performs Git operations, checks dispatch, package and repository permissions and branch rules, and allows branch creation or fast-forward pushes. A push succeeds only after the executor reads the target branch back and matches its SHA to the delivery. Patches, the answer and `deliverables/runtime-source-revision.json` are uploaded with the work item. Repositories are pushed individually, so a successful earlier push remains in place if a later one fails.

Pausing stops the Codex turn before capturing the Git baseline, unpushed local commits and Git-trackable working changes. A recovery dispatch verifies the checkpoint and reconstructs binary changes, new files and deletions before it starts Codex. The platform accepts up to 50 MiB per checkpoint or patch. When a download link expires, the executor requests a fresh link using the checkpoint sequence number, if supplied. If the download still fails, it can use a local snapshot whose SHA-256 matches a digest supplied by the platform. A repeated pause request receives the same pause receipt.

```sh
go run ./cmd/autowonder-switch -url https://autowonder.example -executor-id 42 -repo project=/absolute/path/to/project
```

After a work item closes, the platform can request workspace cleanup. Switch removes the working tree and local Git copies of finished dispatches whose packages predate the closing event. It keeps the final result and checkpoint for redelivery and recovery.

## Conversations and tools

A conversation capability package can bind repositories to a turn. Switch checks out eager repositories before starting Codex. For a lazy repository, Codex calls `executor.open_repository` with its name and receives a local path. A second call returns the same checkout. The executor removes these turn-specific copies when the turn finishes.

Conversations retain the native Codex thread ID for later turns and support cancellation. The executor sends running status, message text and the final reply. With V1 receipts it persists unconfirmed frames.

AutoWonder 1.3.0 queries idle executors registered with the client kind `QODER_CLI` or `QODER_CN_CLI` for their model catalogs. When registering Codex, select the client kind for the catalog you want to populate. The executor uses `QODER_MODEL_CATALOG_V1` to return the models available in local Codex. [How the platform selects an executor](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ProviderModelCatalogService.java)

For work items with a frozen PRD, Codex can use `executor.report_acceptance` to report results against the requirement keys in `design/requirements.json`. A passing verdict requires evidence. A not-applicable verdict requires a reason.

When a dispatch or conversation supplies a platform tool token, Codex can use the `autowonder` tool to list and call the authorized MCP tools. The executor uses the token to contact the platform. The supplied environment variables reach the current turn through the Codex command environment settings.

Task and conversation capability packages can declare third-party MCP servers in `skills.json`. Switch supports stdio, HTTP and SSE and resolves environment and header references from the current dispatch’s `mcpSecrets`. Codex calls tools by `server/tool`. If a required server cannot connect, the turn fails. A platform MCP connection test initializes the server and returns its tool catalog.

Skills listed in the package keep their directory layout. Codex reads the skill’s `SKILL.md` and can open its scripts and binary resources for the current dispatch or conversation turn.

For trusted packages, start Switch with `-native-package-plugins` to let Codex load native plugins, including their executable components. The executor selects descriptors listing `codex` in `providers` and stages each bundle in a temporary namespace under the local Codex plugin cache. The current task or conversation turn activates those IDs. When execution returns, Switch removes the temporary files. With the flag off, a required Codex plugin fails the task or turn and an optional one is skipped.

## Attachments, deliverables and human decisions

Files under `refs/` in a task package retain their original bytes in the workspace. In a work-item clarification conversation, the platform includes the file names, byte counts and signed URLs in the turn content. Repeat `-attachment-origin` for the object storage origins you use. The executor downloads from those origins, checks every byte count and makes the batch available in the conversation workspace for this and later turns. PNG, JPEG and WebP files also reach Codex as image inputs. Reading other formats such as PDFs or archives depends on the tools installed on the host.

To deliver conversation files, set `-artifact-listen` to the listener address and `-artifact-url` to an HTTP(S) address reachable by readers. Forward its `/artifacts/` path to the listener. Regular files in the turn output directory and its subdirectories are snapshotted and linked from the final reply. Anyone with a link can download that file for 24 hours without a platform login. The files remain in `published/` in the state directory.

```sh
go run ./cmd/autowonder-switch -url https://autowonder.example -executor-id 42 \
  -attachment-origin https://storage.example \
  -artifact-listen 127.0.0.1:8788 -artifact-url https://files.example
```

When Codex asks a follow-up question, the executor opens a platform form and delivers the answer to the same Codex turn. V1 receipts distinguish receipt of the answer from a successful write. To let conversation users decide native command, file-change or additional-permission requests, enable `-interactive-approvals`. The form shows the complete command or file-change request. Accepting it authorizes that operation once. For an additional-permission request, the form shows the permissions Codex can use for the remainder of the current turn. Questions asking for secrets are excluded from the platform conversation history.

## Current limits

Code checkpoints require a materializable base commit. Git snapshots omit ignored build outputs and files outside the repository.

The conversation command probe returns an empty list. Switch runs Codex app-server, which has no entry point for the interactive terminal’s slash commands.

After a crash, the executor can look up completed turns in local Codex history. Recovering interrupted commands, file changes or approvals depends on the available checkpoint. A session ID alone cannot establish which operations took effect.

A standalone version update supplies the target version to record. AutoWonder also has a release installation and restart operation (`EXECUTOR_UPDATE_RESTART_V1`). Switch does not advertise it.
