package crafting

import (
	"codex/pkg/storage"
	"encoding/json"
	"fmt"
	"sync"
)

// --------- Core Types ----------

type Requirement struct {
	ID  string `json:"id"`
	Qty int    `json:"qty"`
}

type Craftable struct {
	ID           string        `json:"id"`
	Requirements []Requirement `json:"requirements"`
}

// --------- Manager ----------

type Manager struct {
	craftables   map[string]Craftable
	requireIndex map[string][]string
}

func init() {
    // Register load and save functions
    storage.SM().BindFuncs("crafting", LoadManagers, nil)
}

func LoadManagers(data json.RawMessage) error {
    var loadedManagers  map[string][]Craftable
    if err := json.Unmarshal(data, &loadedManagers); err != nil {
        return fmt.Errorf("failed to unmarshal crafting: %w", err)
    }
    
    for k, v := range loadedManagers {
        m := NewManager()
		for _, c := range v {
			m.craftables[c.ID] = c
			for _, req := range c.Requirements {
				m.requireIndex[req.ID] = append(m.requireIndex[req.ID], c.ID)
			}
		}
		Register(k, m)
    }
    
    return nil
}

func NewManager() *Manager {
	return &Manager{
		craftables:   make(map[string]Craftable),
		requireIndex: make(map[string][]string),
	}
}

// Forward lookup
func (m *Manager) GetCraftable(id string) (Craftable, bool) {
	c, ok := m.craftables[id]
	return c, ok
}

// Reverse lookup
func (m *Manager) FindByRequirement(reqID string) []Craftable {
	var results []Craftable
	for _, cid := range m.requireIndex[reqID] {
		if c, ok := m.craftables[cid]; ok {
			results = append(results, c)
		}
	}
	return results
}

// --------- Global Registry ----------

var (
	registry = make(map[string]*Manager)
	mu       sync.RWMutex
)

// Register new manager under a namespace
func Register(name string, manager *Manager) int {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = manager
	return 1
}

// Get manager by namespace
func Get(name string) (*Manager, bool) {
	mu.RLock()
	defer mu.RUnlock()
	m, ok := registry[name]
	return m, ok
}

// Reset a single manager
func Reset(name string) {
	mu.Lock()
	defer mu.Unlock()
	delete(registry, name)
}

// Reset all managers
func ResetAll() {
	mu.Lock()
	defer mu.Unlock()
	registry = make(map[string]*Manager)
}
