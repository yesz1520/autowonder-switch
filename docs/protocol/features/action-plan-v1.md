# Action plan confirmation

English | [中文](action-plan-v1.zh.md)

Heartbeat capability identifier: `ACTION_PLAN_V1`.

The platform assistant lists the tools available to its conversation token. For writes that require confirmation, call `autowonder.propose_platform_actions` with 1 to 10 `{tool, arguments}` objects in `actions`. The platform freezes the arguments and returns the plan ID, hash and status. The executor reports the pending plan to the user. Once the conversation owner confirms in the platform UI, the server executes the frozen actions.

## Example declaration

```json
{
  "type": "HEARTBEAT",
  "protocolFeatures": [
    "ACTION_PLAN_V1"
  ]
}
```

## Related protocol

- [CONVERSATION_TURN](../messages/conversation-turn.md)

Capability implementation in [ConversationProtocolFeatures.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/ConversationProtocolFeatures.java) and [PlatformConversationService.java](https://github.com/aliyun/alibabacloud-autowonder/blob/0a585550f1d2369f73b2b8b620c206d8f5ec1e57/src/main/java/com/aliyun/autowonder/conversation/PlatformConversationService.java).
