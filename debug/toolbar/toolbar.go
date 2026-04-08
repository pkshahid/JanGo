package toolbar

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Toolbar represents a single request's debug toolbar instance.
type Toolbar struct {
	ID        string
	Panels    map[string]Panel
	Ordered   []string // order of panel names
	CreatedAt time.Time
}

var (
	storeMu sync.RWMutex
	store   = make(map[string]*Toolbar)
)

// NewToolbar creates a new Toolbar instance with a unique ID.
func NewToolbar() *Toolbar {
	id := uuid.New().String()
	tb := &Toolbar{
		ID:        id,
		Panels:    make(map[string]Panel),
		CreatedAt: time.Now(),
	}
	return tb
}

// AddPanel adds a panel to the toolbar.
func (tb *Toolbar) AddPanel(p Panel) {
	name := p.Name()
	tb.Panels[name] = p
	tb.Ordered = append(tb.Ordered, name)
}

// Save stores the toolbar in the global store.
func (tb *Toolbar) Save() {
	storeMu.Lock()
	defer storeMu.Unlock()
	store[tb.ID] = tb

	// Simple cleanup of old toolbars to prevent memory leaks in dev
	if len(store) > 100 {
		var oldest string
		var oldestTime time.Time
		first := true
		for id, t := range store {
			if first || t.CreatedAt.Before(oldestTime) {
				oldest = id
				oldestTime = t.CreatedAt
				first = false
			}
		}
		delete(store, oldest)
	}
}

// GetToolbar retrieves a toolbar by ID.
func GetToolbar(id string) *Toolbar {
	storeMu.RLock()
	defer storeMu.RUnlock()
	return store[id]
}
