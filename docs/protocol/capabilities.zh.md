# 能力与控制消息

[English](capabilities.md) | 中文

以下列出社区版 1.3.0 执行器协议的能力标识和 WebSocket 消息。

## 心跳中的能力标识

在每次心跳的 `protocolFeatures` 数组中填写支持的标识，保留下表中的大小写。

| 能力标识 | 用途 |
|---|---|
| [dispatch_inventory_v1](features/dispatch-inventory-v1.zh.md) | 派单持有清单 |
| [CONVERSATION_COMPUTE_INVENTORY_V1](features/conversation-compute-inventory-v1.zh.md) | 轮次计算清单 |
| [TASK_MODEL_SELECTION_V1](features/task-model-selection-v1.zh.md) | 派单模型选择 |
| [TASK_PACKAGE_SIGNATURE_V1](features/task-package-signature-v1.zh.md) | 任务包验签 |
| [TASK_PACKAGE_HOOKS_V1](features/task-package-hooks-v1.zh.md) | 任务包 hook |
| [TASK_PACKAGE_TOOL_HOOKS_V1](features/task-package-tool-hooks-v1.zh.md) | 阻断式工具 hook |
| [AGENT_ENVIRONMENT_VARIABLES_V1](features/agent-environment-variables-v1.zh.md) | 执行环境变量 |
| [DEBUG_LOG_V1](features/debug-log-v1.zh.md) | 派单调试日志 |
| [DISPATCH_REQUESTS_V1](features/dispatch-requests-v1.zh.md) | 派单中的人工请求 |
| [CHECKPOINT_LIGHTWEIGHT_V1](features/checkpoint-lightweight-v1.zh.md) | 轻量检查点恢复 |
| [terminal_steer_v1](features/terminal-steer-v1.zh.md) | 运行中追加指令 |
| [CONVERSATION_TURN_EVENT](features/conversation-turn-event.zh.md) | 轮次事件 |
| [CONVERSATION_TURN_CANCEL](features/conversation-turn-cancel.zh.md) | 取消会话轮次 |
| [CONVERSATION_ACP_INTERACTION_V1](features/conversation-acp-interaction-v1.zh.md) | 轮次内的 ACP 交互 |
| [CONVERSATION_DELIVERY_ACK_V1](features/conversation-delivery-ack-v1.zh.md) | 轮次事件与结果回执 |
| [CONVERSATION_ANSWER_DELIVERY_V1](features/conversation-answer-delivery-v1.zh.md) | 交互答复投递 |
| [ATTACHMENT_MANIFEST_V1](features/attachment-manifest-v1.zh.md) | 会话附件清单 |
| [ARTIFACT_OUTPUT_V1](features/artifact-output-v1.zh.md) | 会话产物输出 |
| [ACTION_PLAN_V1](features/action-plan-v1.zh.md) | 行动计划确认 |
| [QODER_MODEL_CATALOG_V1](features/qoder-model-catalog-v1.zh.md) | 可用模型目录 |
| [EXECUTOR_RESTART_V1](features/executor-restart-v1.zh.md) | 普通重启 |
| [EXECUTOR_UPDATE_RESTART_V1](features/executor-update-restart-v1.zh.md) | 更新后重启 |
| [EXECUTOR_LAUNCH_COMMAND_V1](features/executor-launch-command-v1.zh.md) | 应用启动配置 |
| [EXECUTOR_RESTART_POLICY_V1](features/executor-restart-policy-v1.zh.md) | 重启时机 |
| [EXECUTOR_UPDATE_V1](features/executor-update-v1.zh.md) | 远程升级 |
| [WORKITEM_WORKSPACE_CLEANUP](features/workitem-workspace-cleanup.zh.md) | 工单工作区清理 |

## WebSocket 消息

| 消息类型 | 方向 | 用途 |
|---|---|---|
| [HEARTBEAT](heartbeat.zh.md) | 执行器 → 平台 | 上报在线状态和执行清单 |
| [TASK_DISPATCH](messages/task-dispatch.zh.md) | 平台 → 执行器 | 下发派单 |
| [TASK_ACK](messages/task-ack.zh.md) | 执行器 → 平台 | 确认接单 |
| [TASK_BUSY](messages/task-busy.zh.md) | 执行器 → 平台 | 报告忙碌 |
| [TASK_PROGRESS](messages/task-progress.zh.md) | 执行器 → 平台 | 上报进度 |
| [TASK_RESULT](messages/task-result.zh.md) | 执行器 → 平台 | 交付结果 |
| [TASK_RESULT_ACK](messages/task-result-ack.zh.md) | 平台 → 执行器 | 确认派单结果 |
| [TASK_PAUSE](messages/task-pause.zh.md) | 平台 → 执行器 | 暂停或取消派单 |
| [TASK_PAUSED](messages/task-paused.zh.md) | 执行器 → 平台 | 报告已暂停 |
| [TASK_PAUSE_FAILED](messages/task-pause-failed.zh.md) | 执行器 → 平台 | 报告暂停失败 |
| [TASK_HANDOFF](messages/task-handoff.zh.md) | 执行器 → 平台 | 请求交接 |
| [TASK_HANDOFF_RESULT](messages/task-handoff-result.zh.md) | 平台 → 执行器 | 返回交接决定 |
| [TASK_GUIDANCE](messages/task-guidance.zh.md) | 平台 → 执行器 | 下发任务指导 |
| [TASK_GUIDANCE_ACK](messages/task-guidance-ack.zh.md) | 执行器 → 平台 | 回复任务指导 |
| [TASK_STEER](messages/task-steer.zh.md) | 平台 → 执行器 | 向运行中的任务追加指令 |
| [TASK_STEER_ACK](messages/task-steer-ack.zh.md) | 执行器 → 平台 | 确认追加指令 |
| [ARTIFACT_UPLOADED](messages/artifact-uploaded.zh.md) | 执行器 → 平台 | 登记已上传产物 |
| [CONVERSATION_TURN](messages/conversation-turn.zh.md) | 平台 → 执行器 | 派发会话轮次 |
| [CONVERSATION_TURN_EVENT](messages/conversation-turn-event.zh.md) | 执行器 → 平台 | 上报轮次事件 |
| [CONVERSATION_TURN_CANCEL](messages/conversation-turn-cancel.zh.md) | 平台 → 执行器 | 下发轮次取消 |
| [CONVERSATION_TURN_ACK](messages/conversation-turn-ack.zh.md) | 执行器 → 平台 | 提交轮次最终回复 |
| [CONVERSATION_DELIVERY_ACK](messages/conversation-delivery-ack.zh.md) | 平台 → 执行器 | 确认收到轮次消息 |
| [CONVERSATION_COMMANDS_PROBE](messages/conversation-commands-probe.zh.md) | 平台 → 执行器 | 查询会话命令 |
| [CONVERSATION_COMMANDS_RESULT](messages/conversation-commands-result.zh.md) | 执行器 → 平台 | 返回会话命令 |
| [CONVERSATION_ELICITATION_REPLY](messages/conversation-elicitation-reply.zh.md) | 平台 → 执行器 | 投递交互答复 |
| [CONVERSATION_ELICITATION_REPLY_ACK](messages/conversation-elicitation-reply-ack.zh.md) | 执行器 → 平台 | 确认交互答复 |
| [QODER_MODEL_CATALOG_REQUEST](messages/qoder-model-catalog-request.zh.md) | 平台 → 执行器 | 查询模型目录 |
| [QODER_MODEL_CATALOG_RESULT](messages/qoder-model-catalog-result.zh.md) | 执行器 → 平台 | 返回模型目录 |
| [MCP_CONNECTION_TEST](messages/mcp-connection-test.zh.md) | 平台 → 执行器 | 测试 MCP 连接 |
| [MCP_CONNECTION_TEST_RESULT](messages/mcp-connection-test-result.zh.md) | 执行器 → 平台 | 返回 MCP 测试结果 |
| [WORKITEM_WORKSPACE_CLEANUP](messages/workitem-workspace-cleanup.zh.md) | 平台 → 执行器 | 清理指定工作区 |
| [EXECUTOR_RESTART](messages/executor-restart.zh.md) | 平台 → 执行器 | 请求重启 |
| [EXECUTOR_RESTART_RESULT](messages/executor-restart-result.zh.md) | 执行器 → 平台 | 上报重启进度 |
| [EXECUTOR_UPGRADE](messages/executor-upgrade.zh.md) | 平台 → 执行器 | 请求升级 |
| [EXECUTOR_UPGRADE_RESULT](messages/executor-upgrade-result.zh.md) | 执行器 → 平台 | 上报升级进度或结果 |
| [EXECUTOR_UPGRADE_RESULT_SUPPORT](messages/executor-upgrade-result-support.zh.md) | 平台 → 执行器 | 确认支持升级结果回执 |
| [EXECUTOR_UPGRADE_RESULT_ACK](messages/executor-upgrade-result-ack.zh.md) | 平台 → 执行器 | 确认收到升级结果 |
