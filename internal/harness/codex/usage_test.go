package codex

import (
	"encoding/json"
	"testing"

	"github.com/yesz1520/autowonder-switch/internal/harness"
)

func TestUsageSeparatesCacheAndReasoning(t *testing.T) {
	var update tokenUpdate
	err := json.Unmarshal([]byte(`{"last":{"inputTokens":100,"cachedInputTokens":60,"cacheWriteInputTokens":5,"outputTokens":30,"reasoningOutputTokens":20},"total":{"totalTokens":25000}}`), &update)
	if err != nil {
		t.Fatal(err)
	}
	var usage harness.Usage
	update.addLast(&usage)
	if usage.Input != 35 || usage.CacheRead != 60 || usage.CacheWrite != 5 || usage.Output != 10 || usage.Reasoning != 20 {
		t.Fatalf("unexpected token categories: %+v", usage)
	}
}
