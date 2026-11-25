package daemon

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

const journalBytes = 512 << 10
const recentEvents = 200

func (m *Manager) eventPath(p Profile) string {
	identity := fmt.Sprintf("%s/%s/%d", p.Key, p.URL, p.ID)
	return filepath.Join(filepath.Dir(m.file), fmt.Sprintf("events-%x.jsonl", sha256.Sum256([]byte(identity))))
}

func (m *Manager) recordEvent(p Profile, e aw.Event) {
	m.journalMu.Lock()
	defer m.journalMu.Unlock()
	if err := appendEvent(m.eventPath(p), e); err != nil {
		log.Printf("executor %d: append event: %v", p.ID, err)
	}
}

func appendEvent(path string, e aw.Event) error {
	if st, err := os.Stat(path); err == nil && st.Size() >= journalBytes {
		if err = os.Rename(path, path+".1"); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(e)
	closeErr := f.Close()
	return errors.Join(err, closeErr)
}

func recent(path string) ([]aw.Event, error) {
	rows := make([]aw.Event, 0, recentEvents)
	for _, name := range []string{path, path + ".1"} {
		b, err := os.ReadFile(name)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		// Walk backwards so opening the tab reads at most the latest 200 entries.
		for len(b) > 0 && len(rows) < recentEvents {
			b = bytes.TrimRight(b, "\n")
			i := bytes.LastIndexByte(b, '\n')
			var e aw.Event
			if json.Unmarshal(b[i+1:], &e) == nil {
				rows = append(rows, e)
			}
			if i < 0 {
				break
			}
			b = b[:i]
		}
		if len(rows) == recentEvents {
			break
		}
	}
	return rows, nil
}

func (m *Manager) eventsHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	idx := slices.IndexFunc(m.profiles, func(p Profile) bool { return p.Key == r.PathValue("key") })
	if idx < 0 {
		m.mu.Unlock()
		http.NotFound(w, r)
		return
	}
	path := m.eventPath(m.profiles[idx])
	m.mu.Unlock()
	m.journalMu.Lock()
	rows, err := recent(path)
	m.journalMu.Unlock()
	if err != nil {
		http.Error(w, "无法读取执行记录", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rows)
}
