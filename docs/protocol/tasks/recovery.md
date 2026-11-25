# Pause, cancel and recover

English | [中文](recovery.zh.md)

Save a recovery checkpoint on pause. An executor can reclaim its dispatch after reconnecting, or resume from a checkpoint supplied on a subsequent dispatch.

## Interfaces and events

- [Pause or cancel a dispatch](../messages/task-pause.md)
- [Confirm a pause](../messages/task-paused.md)
- [Report a failed pause](../messages/task-pause-failed.md)
- [Reclaim an active dispatch](../http/recovery-claim.md)
- [Refresh a checkpoint URL](../http/checkpoint-url.md)
- [TASK_DISPATCH](../messages/task-dispatch.md)
