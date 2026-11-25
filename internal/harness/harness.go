// Package harness defines the local execution contract independently of AutoWonder.
package harness

import "context"

type Model struct{ ID, Name string }
type Catalog struct {
	Models  []Model
	Default string
}
type PlatformTool func(context.Context, string, map[string]any) (string, error)
type Answer struct {
	Action  string
	Content map[string]any
	Written func(bool)
}
type Ask func(context.Context, string, map[string]any) (Answer, error)
type Request struct {
	ReadOnlyPaths                                   []string // Executor-owned repository metadata, conversation checkouts and packaged skills.
	MemoryMode                                      string
	PluginIDs                                       []string
	ToolHookCommands                                map[string]string
	OutputSchema                                    map[string]any
	Ask                                             Ask
	Images                                          []string
	Approvals                                       bool
	Platform                                        PlatformTool
	Environment                                     map[string]string
	Readable, Writable, Fork                        bool
	Directory, Prompt, Instructions, Model, Session string
}

// Usage splits input and output into exclusive categories so their sum is the turn total.
type Usage struct {
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	Input      int64  `json:"inputTokens"`
	Output     int64  `json:"outputTokens"`
	CacheRead  int64  `json:"cacheReadTokens"`
	CacheWrite int64  `json:"cacheWriteTokens"`
	Reasoning  int64  `json:"reasoningTokens"`
}
type Reply struct {
	Text, Session, Model string
	Usage                *Usage
}
type Hooks struct {
	Session func(string) error
	// TurnStarted records the native identifiers before the adapter accepts control messages.
	TurnStarted func(session, turn string) error
	// Ready installs a control for this turn only.
	Ready   func(steer func(context.Context, string) error)
	Message func(string)
}
type Backend interface {
	Version(context.Context) (string, error)
	Models(context.Context) (Catalog, error)
	Run(context.Context, Request, Hooks) (Reply, error)
}

// TurnReader inspects a previous invocation without starting model work.
type TurnReader interface {
	ReadTurn(ctx context.Context, directory, session, turn string) (Reply, string, error)
}
