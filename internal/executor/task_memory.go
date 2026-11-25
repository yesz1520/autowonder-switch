package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
)

const memoryTool = "executor.workitem_memory"
const memorySchema = `{"name":"executor.workitem_memory","description":"Access AutoWonder memory for this dispatch. list returns the index, paths and document versions; read returns one current document and records its recall. changes reads a page after a store revision. For create, update and delete use one idempotencyKey per logical write and reuse it on retry. Update and delete require the version you read. Use newPath to rename in an update.","inputSchema":{"type":"object","properties":{"operation":{"type":"string","enum":["list","read","changes","create","update","delete"]},"storeId":{"type":"integer"},"path":{"type":"string"},"newPath":{"type":"string"},"contentMd":{"type":"string"},"expectedVersion":{"type":"integer"},"idempotencyKey":{"type":"string"},"afterRevision":{"type":"integer"}},"required":["operation"],"additionalProperties":false}}`

func (r *runtime) bindTaskMemory(q *aw.TaskDispatch, spec *harness.Request) error {
	if r.launch.MemoryMode == "none" || r.launch.MemoryMode == "provider-local" || q.MemorySnapshotPath == "" || q.DispatchMCPToken == "" {
		return nil
	}
	m, err := r.client.Memory(q.DispatchID, q.DispatchMCPToken)
	if err != nil {
		return err
	}
	platform := spec.Platform
	spec.Instructions += "\nUse executor.workitem_memory to look up relevant shared memory. Consult its index before reading individual documents."
	spec.Platform = func(ctx context.Context, name string, args map[string]any) (string, error) {
		if name == "" {
			return appendTool(ctx, platform, json.RawMessage(memorySchema))
		}
		if name != memoryTool {
			if platform == nil {
				return "", errors.New("unknown platform tool")
			}
			return platform(ctx, name, args)
		}
		data, err := json.Marshal(args)
		if err != nil {
			return "", err
		}
		var in struct {
			aw.MemoryMutationRequest
			StoreID       int64
			AfterRevision int64
		}
		if err = json.Unmarshal(data, &in); err != nil {
			return "", err
		}
		if in.Operation != "list" && in.StoreID <= 0 {
			return "", errors.New("storeId required")
		}
		var result any
		switch in.Operation {
		case "list":
			snapshot, e := m.Snapshot(ctx)
			if e != nil {
				return "", e
			}
			for _, store := range snapshot.Stores {
				for i := range store.Documents {
					store.Documents[i].ContentMD = ""
				}
			}
			result = snapshot
		case "read":
			result, err = readTaskMemory(ctx, m, in.StoreID, in.Path, q.MemoryRecallPath != "")
		case "changes":
			if q.MemoryChangesPath == "" || in.AfterRevision < 0 {
				return "", errors.New("changes unavailable or invalid afterRevision")
			}
			result, err = m.Changes(ctx, in.StoreID, in.AfterRevision, 100)
		case "create", "update", "delete":
			if q.MemoryMutationPath == "" || in.Path == "" || strings.TrimSpace(in.IdempotencyKey) == "" {
				return "", errors.New("write requires mutations, path and idempotencyKey")
			}
			if in.Operation != "create" && in.ExpectedVersion == nil {
				return "", errors.New("expectedVersion required")
			}
			in.Operation = strings.ToUpper(in.Operation)
			if in.NewPath == "" {
				in.NewPath = in.Path
			}
			in.MaintenanceLeaseID = ""
			result, err = m.Mutate(ctx, in.StoreID, in.MemoryMutationRequest)
		default:
			return "", fmt.Errorf("unknown memory operation %q", in.Operation)
		}
		if err != nil {
			return "", err
		}
		b, err := json.Marshal(result)
		return string(b), err
	}
	return nil
}

func readTaskMemory(ctx context.Context, m *aw.Memory, storeID int64, path string, recall bool) (any, error) {
	snapshot, err := m.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	for _, s := range snapshot.Stores {
		if s.ID != storeID {
			continue
		}
		for _, d := range s.Documents {
			if d.Path != path {
				continue
			}
			if recall {
				if err := m.Recall(ctx, storeID, path); err != nil {
					return nil, err
				}
			}
			return d, nil
		}
	}
	return nil, errors.New("memory document not found")
}
