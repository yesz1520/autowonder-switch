# Capabilities and controls

English | [中文](capabilities.zh.md)

This index covers executor capabilities and WebSocket messages in Community 1.3.0.

## Heartbeat capabilities

Include supported identifiers in the `protocolFeatures` array in every heartbeat. Preserve the spelling and case shown below.

| Feature | Purpose |
|---|---|
| [dispatch_inventory_v1](features/dispatch-inventory-v1.md) | Dispatch ownership inventory |
| [CONVERSATION_COMPUTE_INVENTORY_V1](features/conversation-compute-inventory-v1.md) | Computing turn inventory |
| [TASK_MODEL_SELECTION_V1](features/task-model-selection-v1.md) | Per-dispatch model selection |
| [TASK_PACKAGE_SIGNATURE_V1](features/task-package-signature-v1.md) | Task package signatures |
| [TASK_PACKAGE_HOOKS_V1](features/task-package-hooks-v1.md) | Task package hooks |
| [TASK_PACKAGE_TOOL_HOOKS_V1](features/task-package-tool-hooks-v1.md) | Blocking tool hooks |
| [AGENT_ENVIRONMENT_VARIABLES_V1](features/agent-environment-variables-v1.md) | Execution environment variables |
| [DEBUG_LOG_V1](features/debug-log-v1.md) | Dispatch debug logs |
| [DISPATCH_REQUESTS_V1](features/dispatch-requests-v1.md) | Human requests during dispatches |
| [CHECKPOINT_LIGHTWEIGHT_V1](features/checkpoint-lightweight-v1.md) | Lightweight checkpoint recovery |
| [terminal_steer_v1](features/terminal-steer-v1.md) | Steering a running dispatch |
| [CONVERSATION_TURN_EVENT](features/conversation-turn-event.md) | Turn events |
| [CONVERSATION_TURN_CANCEL](features/conversation-turn-cancel.md) | Turn cancellation |
| [CONVERSATION_ACP_INTERACTION_V1](features/conversation-acp-interaction-v1.md) | ACP interactions during a turn |
| [CONVERSATION_DELIVERY_ACK_V1](features/conversation-delivery-ack-v1.md) | Turn event and result receipts |
| [CONVERSATION_ANSWER_DELIVERY_V1](features/conversation-answer-delivery-v1.md) | Interactive answer delivery |
| [ATTACHMENT_MANIFEST_V1](features/attachment-manifest-v1.md) | Conversation attachment manifests |
| [ARTIFACT_OUTPUT_V1](features/artifact-output-v1.md) | Conversation output artifacts |
| [ACTION_PLAN_V1](features/action-plan-v1.md) | Action plan confirmation |
| [QODER_MODEL_CATALOG_V1](features/qoder-model-catalog-v1.md) | Available model catalog |
| [EXECUTOR_RESTART_V1](features/executor-restart-v1.md) | Restart |
| [EXECUTOR_UPDATE_RESTART_V1](features/executor-update-restart-v1.md) | Update and restart |
| [EXECUTOR_LAUNCH_COMMAND_V1](features/executor-launch-command-v1.md) | Restart with launch configuration |
| [EXECUTOR_RESTART_POLICY_V1](features/executor-restart-policy-v1.md) | Restart policy |
| [EXECUTOR_UPDATE_V1](features/executor-update-v1.md) | Remote upgrade |
| [WORKITEM_WORKSPACE_CLEANUP](features/workitem-workspace-cleanup.md) | Work item workspace cleanup |

## WebSocket messages

| Message type | Direction | Purpose |
|---|---|---|
| [HEARTBEAT](heartbeat.md) | Executor → platform | Report presence and execution inventory |
| [TASK_DISPATCH](messages/task-dispatch.md) | Platform → executor | Dispatch a task |
| [TASK_ACK](messages/task-ack.md) | Executor → platform | Accept a dispatch |
| [TASK_BUSY](messages/task-busy.md) | Executor → platform | Report a busy executor |
| [TASK_PROGRESS](messages/task-progress.md) | Executor → platform | Report progress |
| [TASK_RESULT](messages/task-result.md) | Executor → platform | Deliver a result |
| [TASK_RESULT_ACK](messages/task-result-ack.md) | Platform → executor | Acknowledge a dispatch outcome |
| [TASK_PAUSE](messages/task-pause.md) | Platform → executor | Pause or cancel a dispatch |
| [TASK_PAUSED](messages/task-paused.md) | Executor → platform | Confirm a pause |
| [TASK_PAUSE_FAILED](messages/task-pause-failed.md) | Executor → platform | Report a failed pause |
| [TASK_HANDOFF](messages/task-handoff.md) | Executor → platform | Request a handoff |
| [TASK_HANDOFF_RESULT](messages/task-handoff-result.md) | Platform → executor | Return a handoff decision |
| [TASK_GUIDANCE](messages/task-guidance.md) | Platform → executor | Send task guidance |
| [TASK_GUIDANCE_ACK](messages/task-guidance-ack.md) | Executor → platform | Acknowledge task guidance |
| [TASK_STEER](messages/task-steer.md) | Platform → executor | Steer a running task |
| [TASK_STEER_ACK](messages/task-steer-ack.md) | Executor → platform | Acknowledge steering |
| [ARTIFACT_UPLOADED](messages/artifact-uploaded.md) | Executor → platform | Register an uploaded artifact |
| [CONVERSATION_TURN](messages/conversation-turn.md) | Platform → executor | Dispatch a turn |
| [CONVERSATION_TURN_EVENT](messages/conversation-turn-event.md) | Executor → platform | Report a turn event |
| [CONVERSATION_TURN_CANCEL](messages/conversation-turn-cancel.md) | Platform → executor | Cancel a turn |
| [CONVERSATION_TURN_ACK](messages/conversation-turn-ack.md) | Executor → platform | Submit a final turn reply |
| [CONVERSATION_DELIVERY_ACK](messages/conversation-delivery-ack.md) | Platform → executor | Acknowledge turn delivery |
| [CONVERSATION_COMMANDS_PROBE](messages/conversation-commands-probe.md) | Platform → executor | Probe conversation commands |
| [CONVERSATION_COMMANDS_RESULT](messages/conversation-commands-result.md) | Executor → platform | Return conversation commands |
| [CONVERSATION_ELICITATION_REPLY](messages/conversation-elicitation-reply.md) | Platform → executor | Deliver an interactive answer |
| [CONVERSATION_ELICITATION_REPLY_ACK](messages/conversation-elicitation-reply-ack.md) | Executor → platform | Acknowledge an interactive answer |
| [QODER_MODEL_CATALOG_REQUEST](messages/qoder-model-catalog-request.md) | Platform → executor | Request the model catalog |
| [QODER_MODEL_CATALOG_RESULT](messages/qoder-model-catalog-result.md) | Executor → platform | Return the model catalog |
| [MCP_CONNECTION_TEST](messages/mcp-connection-test.md) | Platform → executor | Test an MCP connection |
| [MCP_CONNECTION_TEST_RESULT](messages/mcp-connection-test-result.md) | Executor → platform | Return an MCP test result |
| [WORKITEM_WORKSPACE_CLEANUP](messages/workitem-workspace-cleanup.md) | Platform → executor | Clean up a workspace |
| [EXECUTOR_RESTART](messages/executor-restart.md) | Platform → executor | Request a restart |
| [EXECUTOR_RESTART_RESULT](messages/executor-restart-result.md) | Executor → platform | Report restart progress |
| [EXECUTOR_UPGRADE](messages/executor-upgrade.md) | Platform → executor | Request an upgrade |
| [EXECUTOR_UPGRADE_RESULT](messages/executor-upgrade-result.md) | Executor → platform | Report upgrade progress or outcome |
| [EXECUTOR_UPGRADE_RESULT_SUPPORT](messages/executor-upgrade-result-support.md) | Platform → executor | Confirm upgrade receipt support |
| [EXECUTOR_UPGRADE_RESULT_ACK](messages/executor-upgrade-result-ack.md) | Platform → executor | Acknowledge an upgrade outcome |
