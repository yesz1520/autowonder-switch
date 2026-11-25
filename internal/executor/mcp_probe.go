package executor

import (
	"context"
	"encoding/json"
	"os"
	"time"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/mcpclient"
)

func (r *runtime) probeMCP(ctx context.Context, s *aw.Session, q *aw.MCPConnectionTest) error {
	if q.TestID == "" || (q.ExecutorID != 0 && q.ExecutorID != r.executorID) {
		return s.Send(ctx, &aw.MCPConnectionTestResult{TestID: q.TestID, Message: "invalid probe identity"})
	}
	r.wg.Go(func() {
		start := time.Now()
		reply := aw.MCPConnectionTestResult{TestID: q.TestID}
		work, err := os.MkdirTemp(r.dir, ".mcp-probe-")
		if err == nil {
			defer os.RemoveAll(work)
			cfg := mcpclient.Server{Transport: q.Transport, Command: q.Command, Args: q.Args, Env: q.Env, URL: q.URL, Headers: q.Headers, TimeoutSeconds: q.TimeoutSeconds}
			session, e := mcpclient.Open(ctx, cfg, nil, work, r.codexBinary)
			err = e
			if err == nil {
				for _, tool := range session.Tools {
					raw, e := json.Marshal(tool)
					if e != nil {
						err = e
						break
					}
					reply.Tools = append(reply.Tools, raw)
				}
				session.Close()
			}
		}
		reply.Success = err == nil
		if err != nil {
			reply.Message = err.Error()
		}
		reply.DurationMS = time.Since(start).Milliseconds()
		if ctx.Err() == nil {
			_ = s.Send(ctx, &reply)
		}
	})
	return nil
}
