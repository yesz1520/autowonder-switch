package autowonder

type Feature string

const (
	FeatureDispatchInventoryV1            Feature = "dispatch_inventory_v1"
	FeatureConversationComputeInventoryV1 Feature = "CONVERSATION_COMPUTE_INVENTORY_V1"
	FeatureTaskModelSelectionV1           Feature = "TASK_MODEL_SELECTION_V1"
	FeatureTaskPackageSignatureV1         Feature = "TASK_PACKAGE_SIGNATURE_V1"
	FeatureTaskPackageHooksV1             Feature = "TASK_PACKAGE_HOOKS_V1"
	FeatureTaskPackageToolHooksV1         Feature = "TASK_PACKAGE_TOOL_HOOKS_V1"
	FeatureAgentEnvironmentVariablesV1    Feature = "AGENT_ENVIRONMENT_VARIABLES_V1"
	FeatureDebugLogV1                     Feature = "DEBUG_LOG_V1"
	FeatureDispatchRequestsV1             Feature = "DISPATCH_REQUESTS_V1"
	FeatureCheckpointLightweightV1        Feature = "CHECKPOINT_LIGHTWEIGHT_V1"
	FeatureTerminalSteerV1                Feature = "terminal_steer_v1"
	FeatureConversationTurnEvent          Feature = "CONVERSATION_TURN_EVENT"
	FeatureConversationTurnCancel         Feature = "CONVERSATION_TURN_CANCEL"
	FeatureConversationAcpInteractionV1   Feature = "CONVERSATION_ACP_INTERACTION_V1"
	FeatureConversationDeliveryAckV1      Feature = "CONVERSATION_DELIVERY_ACK_V1"
	FeatureConversationAnswerDeliveryV1   Feature = "CONVERSATION_ANSWER_DELIVERY_V1"
	FeatureAttachmentManifestV1           Feature = "ATTACHMENT_MANIFEST_V1"
	FeatureArtifactOutputV1               Feature = "ARTIFACT_OUTPUT_V1"
	FeatureActionPlanV1                   Feature = "ACTION_PLAN_V1"
	FeatureQoderModelCatalogV1            Feature = "QODER_MODEL_CATALOG_V1"
	FeatureExecutorRestartV1              Feature = "EXECUTOR_RESTART_V1"
	FeatureExecutorUpdateRestartV1        Feature = "EXECUTOR_UPDATE_RESTART_V1"
	FeatureExecutorLaunchCommandV1        Feature = "EXECUTOR_LAUNCH_COMMAND_V1"
	FeatureExecutorRestartPolicyV1        Feature = "EXECUTOR_RESTART_POLICY_V1"
	FeatureExecutorUpdateV1               Feature = "EXECUTOR_UPDATE_V1"
	FeatureWorkitemWorkspaceCleanup       Feature = "WORKITEM_WORKSPACE_CLEANUP"
)
