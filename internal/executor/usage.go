package executor

import (
	"context"
	"log"
	"time"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
)

type usageKey struct{ step, provider, model string }
type taskUsage map[usageKey]harness.Usage

// The platform upserts one row per dispatch, step, provider and model. Send
// cumulative values so a workflow revision and a retried POST both count once.
func (u taskUsage) report(ctx context.Context, client *aw.Client, dispatch int64, step string, turn *harness.Usage) {
	if turn == nil {
		return
	}
	key := usageKey{step, turn.Provider, turn.Model}
	total := u[key]
	total.Input += turn.Input
	total.Output += turn.Output
	total.CacheRead += turn.CacheRead
	total.CacheWrite += turn.CacheWrite
	total.Reasoning += turn.Reasoning
	u[key] = total
	entry := aw.UsageEntry{Provider: key.provider, Model: key.model, StepID: step, InputTokens: &total.Input, OutputTokens: &total.Output, CacheReadTokens: &total.CacheRead, CacheWriteTokens: &total.CacheWrite, ReasoningTokens: &total.Reasoning}
	bounded, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	for i := 0; i < 3; i++ {
		err := client.ReportUsage(bounded, dispatch, []aw.UsageEntry{entry})
		if err == nil {
			return
		}
		if i == 2 || !retryHumanRequest(err) || bounded.Err() != nil {
			log.Printf("dispatch %d step %s usage: %v", dispatch, step, err)
			return
		}
	}
}
