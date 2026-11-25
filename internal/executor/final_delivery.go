package executor

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

// A ready file is both the intent to enqueue and the future duplicate-response
// cache. Promote it only after the outbox can replay the exact payload.
func (r *runtime) finishDelivery(path string, msg aw.Outgoing) {
	if err := r.recordFinal(path, msg); err != nil {
		log.Printf("record final %s: %v", msg.Kind(), err)
	}
	r.report(msg)
}

func (r *runtime) recordFinal(path string, msg aw.Outgoing) error {
	r.finals.Lock()
	defer r.finals.Unlock()
	if err := save(path+".ready", msg); err != nil {
		return err
	}
	return r.enqueueFinal(path, msg)
}

func (r *runtime) enqueueFinal(path string, msg aw.Outgoing) error {
	if err := r.pending.Store(msg); err != nil {
		return err
	}
	if err := os.Rename(path+".ready", path); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func (r *runtime) recoverFinalDeliveries() error {
	for _, pattern := range []string{"task-*/result.json.ready", "conversation-*/turn-*.json.ready"} {
		paths, err := fs.Glob(os.DirFS(r.dir), pattern)
		if err != nil {
			return err
		}
		for _, relative := range paths {
			ready := filepath.Join(r.dir, relative)
			var msg aw.Outgoing
			if filepath.Base(ready) == "result.json.ready" {
				msg = &aw.TaskResult{}
			} else {
				msg = &aw.ConversationTurnAck{}
			}
			ok, err := cached(ready, msg)
			if err != nil {
				return fmt.Errorf("recover %s: %w", ready, err)
			}
			if !ok {
				return errors.New("final delivery disappeared during recovery")
			}
			if err = r.enqueueFinal(strings.TrimSuffix(ready, ".ready"), msg); err != nil {
				return err
			}
		}
	}
	return nil
}

// Redelivery can arrive on a healthy connection after a local storage failure.
func (r *runtime) loadFinal(path string, msg aw.Outgoing) (bool, error) {
	r.finals.Lock()
	defer r.finals.Unlock()
	if ok, err := cached(path, msg); err != nil || ok {
		return ok, err
	}
	ok, err := cached(path+".ready", msg)
	if err != nil || !ok {
		return ok, err
	}
	return true, r.enqueueFinal(path, msg)
}
