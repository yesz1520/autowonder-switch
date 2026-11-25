# 运行 Codex 执行器

[English](codex.md) | 中文

执行器使用本机的 Codex 配置和登录。先列出可用模型。此命令会启动 Codex app-server，但不会连接 AutoWonder 或发起模型轮次。

```sh
go run ./cmd/autowonder-switch -list-models
```

## 连接执行器

在环境变量 `AUTOWONDER_TOKEN` 中设置执行器令牌。把平台地址和执行器 ID 换成自己的值。程序在前台运行，按 Ctrl+C 退出。默认并发容量为 1，技术上限为 50，派单和会话轮次共用容量。

```sh
go run ./cmd/autowonder-switch -url https://autowonder.example -executor-id 42 -capacity 2
```

同一台机器可以用不同的 ID 和对应令牌分别启动进程，共用本机 Codex。状态默认保存在用户配置目录下的 `autowonder-switch` 中，按平台和身份分开。可用 `-state-dir` 指定目录，用 `-codex` 指定可执行文件。在途任务可以跨 WebSocket 重连续跑，进程退出则会中断模型轮次。已保存的最终回复可在重复派单时重新发送。

平台下发升级目标版本时，Switch 将版本号保存在当前执行器身份下，并在后续心跳中上报。重启后仍使用该值。本机 Codex 二进制由运维单独更新。

Linux 和 macOS 上支持远程重启。执行器沿用当前启动参数和环境变量。收到 `WHEN_IDLE` 时停止接单，等在途工单和会话轮次结束后再重启。重启后的首次心跳携带请求 ID。

平台可以向 Linux 和 macOS 执行器下发启动配置。重启后，握手和首次心跳使用新的并发容量，心跳带上请求 ID 和目标运行时版本。新工单和新会话在 `workspaceRoot` 下创建独立目录，已有工作区沿用原路径。`daemonName` 用作日志标识，开启 `debug` 后可在本机日志中看到收到的帧类型。

`memoryMode` 为 `platform` 时，执行器提供平台记忆；为 `provider-local` 时，Codex 沿用本机记忆设置；为 `none` 时，执行器关闭记忆注入。AutoWonder 1.3.0 为 `CODEX_CLI` 生成启动配置时会清空模型、推理强度和上下文窗口。工单单独指定模型时，执行器使用 `modelSelection.model`。

## 执行工单

文本工单读取经过校验的任务包，把回答作为 `result.md` 产物上传，再提交完成检查点和 V2 结果。工单可以指定模型，在运行中追加指令或取消。文本轮次支持暂停，从该执行器已有的 Codex 会话继续或分叉，也可读取本执行器生成的文本检查点。

Codex 返回 token 用量时，执行器按派单、工作流步骤和模型上报。一个步骤退回修改时，该步骤的用量包含修改前后的轮次。会话用量随轮次回执发送，只统计本轮。

工单需要交接时，Codex 可以用 `executor.request_handoff` 指定下一位数字员工或真人。执行器随成功结果提交交接建议，平台决定实际接收方。平台回包保存在派单目录的 `handoff-result.json`。

派单提供记忆接口和派单令牌时，Codex 可以查询记忆库及文档版本，读取文档，或按库的修订号获取后续变更。派单提供召回接口时，读取文档会登记一次召回。写操作带幂等键，更新和删除还会带上读取时的文档版本，平台据此拒绝冲突修改。

如果执行器在保存最终答复后、等待平台确认前退出，重启时会在首次心跳前恢复待确认的工单结果和 V1 会话答复。平台无需重新派单即可收到原答复。

如果执行器在发送结果前退出，平台重发同一次派单尝试时，执行器可以用已保存的工作流答复继续交付。对于无需等待审批或退出 hook 的单轮工单，也可以用会话 ID 和轮次 ID 从本机 Codex 历史中取回已完成的答复。

工作流步骤开启 `askWhenUnclear` 时，Codex 可以在工单里发起表单，收到答复后从同一轮继续。开启 `humanApprovalRequired` 时，执行器在步骤结束后提交审批，附上当前回答和仓库快照的 SHA-256。用户要求修改时，执行器在该步骤的重试次数内将修改意见交回 Codex。通过审批后才推进下一步或交付结果。

派单开启 `debugLog.enabled` 时，执行器按 `maxBytes` 记录阶段时间、派单 ID 和消息字节数。日志压缩后优先上传到签名地址，失败时经平台中转。上传结果随工单结果返回。

任务包可以在仓库准备、Codex 执行、工作流步骤、提交和推送时运行 hook，也可以在失败或清理时运行。执行器先核对 hook 包摘要，再在 Codex 沙箱里按 `failurePolicy` 和 `onceScope` 执行脚本。输出保存在派单目录的 `hooks-<attempt>/logs/<trigger>/<name>/`。

`beforeTool` 和 `afterTool` 分别映射到 Codex 的 `PreToolUse` 和 `PostToolUse`。启动时探测通过才声明工具 hook 能力。前置脚本拒绝可以阻止当次调用；后置脚本收到工具结果时，工具操作已经完成。

### 代码工单

执行器读取 `repos.json`。允许联网时，按仓库的 HTTP(S) 或 SSH 地址检出指定版本，私有仓库使用运行 daemon 的系统用户配置的 Git 凭据或 SSH agent。已有本地副本时可以使用下面的映射；禁止联网时，指定版本必须已在本机。派单固定的 `pinnedBaseCommit` 获取失败时会返回错误。每个派单有独立工作目录，可在里面运行 Git 状态、差异、历史查询和项目测试。

Codex 按工单内容为每个仓库选择补丁、本地提交或推送，执行器负责运行 Git。推送前核对派单、任务包及该仓库的权限和分支规则；只做快进推送或创建分支。推送后查询目标分支，只有 SHA 与本次交付匹配才确认推送成功。补丁、回答和 `deliverables/runtime-source-revision.json` 随工单上传。多个仓库分别推送，前一个仓库已经成功的推送会保留。

暂停代码工单时，先停止 Codex 轮次，再保存仓库基准、未推送的本地提交和 Git 可跟踪的工作区修改。恢复派单从校验过的检查点重建工作目录，包括二进制改动、新文件和删除，再启动 Codex。恢复链接失效时，执行器会按平台提供的检查点序号申请新链接。如果下载仍然失败，且本机快照的 SHA-256 与平台指定的值一致，执行器可以从本机恢复。平台重复请求暂停时，执行器重发同一条暂停回执。平台单个检查点和补丁的大小上限均为 50 MiB。

```sh
go run ./cmd/autowonder-switch -url https://autowonder.example -executor-id 42 -repo project=/absolute/path/to/project
```

工单关闭后，平台可以请求清理工作区。执行器清除关闭事件发生时已有的、已结束派单的工作目录和 Git 副本，并保留最终结果和检查点供重发或恢复。

## 会话和工具

会话能力包可以为本轮指定仓库。执行器在启动 Codex 前检出 `eager` 仓库。对于 `lazy` 仓库，Codex 调用 `executor.open_repository`，传入仓库名即可取得本地路径。同一轮再次打开会返回原路径。轮次结束后，执行器删除本轮的仓库副本。

会话保存原生 Codex thread ID，后续轮次可续聊或取消。执行器发送运行状态、消息正文和最终回复，启用 V1 回执时保存待确认帧。

AutoWonder 1.3.0 只向客户端类型为 `QODER_CLI` 或 `QODER_CN_CLI` 的空闲执行器查询模型目录。在平台上为 Codex 创建执行器时，选择与所需目录对应的类型。执行器通过 `QODER_MODEL_CATALOG_V1` 返回本机 Codex 的可选模型。[平台选择执行器的条件](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/executor/ProviderModelCatalogService.java)

带有冻结版 PRD 的工单可以用 `executor.report_acceptance` 回填逐项验收结果。需求标识取自任务包的 `design/requirements.json`。通过验收时附上证据，标为不适用时填写原因。

派单或会话携带平台工具令牌时，Codex 可通过 `autowonder` 工具列举和调用平台授权的 MCP 工具。令牌由执行器用于请求平台。派发的环境变量通过 Codex 的命令环境配置传入当前轮次。

任务包和会话能力包可在 `skills.json` 中声明第三方 MCP 服务。执行器支持 stdio、HTTP 和 SSE，按本次派发的 `mcpSecrets` 解析环境变量和请求头引用。Codex 用 `服务名/工具名` 调用对应工具。必需服务无法连接时，本轮返回错误。平台的 MCP 连接测试会实际初始化服务并返回工具目录。

执行器按包中的目录结构解开技能。Codex 在本次工单或会话轮次中可以读取技能的 `SKILL.md`，并访问随附的脚本和二进制资源。

如果信任平台下发的插件包，可以用 `-native-package-plugins` 启动执行器，允许 Codex 加载原生插件及其可执行组件。执行器选取 `providers` 包含 `codex` 的插件，将文件解到本机 Codex 插件缓存的临时目录，仅在当前工单或会话轮次中启用。执行结束后删除临时文件。未开启此选项时，必需插件会让本次工单或轮次失败，可选插件会跳过。

## 附件、交付文件和人工确认

工单包中 `refs/` 下的文件按原始字节放入工作目录。工单澄清会话的附件则由平台把文件名、字节数和签名 URL 写入本轮正文。用可重复的 `-attachment-origin` 指定实际对象存储的源地址。执行器只从配置的源下载，整批文件通过字节数校验后才放入本会话工作区，后续轮次可继续读取；PNG、JPEG 和 WebP 还作为图片输入传给 Codex。PDF、压缩包等文件由工作区里的工具读取，可读格式取决于本机安装的软件。

要让会话用户下载文件，用 `-artifact-listen` 指定监听地址，用 `-artifact-url` 指定用户能访问的 HTTP(S) 地址。将该地址的 `/artifacts/` 路径转发到监听端口。Codex 写入本轮输出目录及其子目录的普通文件会保存为快照，最终回复带有下载链接。每个链接 24 小时内有效，持有链接的人可下载文件，无需平台登录。文件保存在状态目录的 `published/` 中。

```sh
go run ./cmd/autowonder-switch -url https://autowonder.example -executor-id 42 \
  -attachment-origin https://storage.example \
  -artifact-listen 127.0.0.1:8788 -artifact-url https://files.example
```

Codex 在会话中发起追问时，执行器发送平台问答卡。答案写回同一个 Codex 轮次，V1 回执区分收到答复和写入成功。需要由会话用户决定本机命令、文件修改或额外权限时，启动参数加 `-interactive-approvals`。表单显示完整的命令或文件修改请求，点击同意只批准这一次操作。对于额外权限请求，表单会列出 Codex 在当前轮次结束前可以使用的权限。需要输入秘密的追问不会写入平台会话记录。

## 当前限制

代码工单的检查点依赖基准提交仍可检出。Git 忽略的构建产物和仓库外文件不在 Git 快照中。

会话命令探针返回空列表。这里运行的是 Codex app-server，交互终端里的斜杠命令没有对应的调用入口。

进程崩溃后，执行器可以查询本机历史里已经完成的轮次。尚未完成的命令、文件修改或审批需要根据已保存的检查点恢复；仅凭会话 ID 无法确认这些操作是否已经生效。

独立的版本更新请求会下发目标版本。AutoWonder 另有一个要求安装发布包再重启的操作（`EXECUTOR_UPDATE_RESTART_V1`），Switch 不声明这项能力。
