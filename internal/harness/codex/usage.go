package codex

import "github.com/yesz1520/autowonder-switch/internal/harness"

type tokenUpdate struct {
	Total struct{ TotalTokens int64 }
	Last  struct {
		InputTokens, OutputTokens, CachedInputTokens, CacheWriteInputTokens, ReasoningOutputTokens int64
	}
}

// Codex includes cached and reasoning tokens in its input and output totals.
// AutoWonder adds the five columns when computing the displayed total.
func (u tokenUpdate) addLast(usage *harness.Usage) {
	v := u.Last
	usage.Input += max(0, v.InputTokens-v.CachedInputTokens-v.CacheWriteInputTokens)
	usage.Output += max(0, v.OutputTokens-v.ReasoningOutputTokens)
	usage.CacheRead += v.CachedInputTokens
	usage.CacheWrite += v.CacheWriteInputTokens
	usage.Reasoning += v.ReasoningOutputTokens
}
