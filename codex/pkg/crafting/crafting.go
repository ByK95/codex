package crafting

import (
	"encoding/json"
	"os"
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

// Load from JSON
func NewManager(path string) (*Manager, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var items []Craftable
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	m := &Manager{
		craftables:   make(map[string]Craftable),
		requireIndex: make(map[string][]string),
	}

	for _, c := range items {
		m.craftables[c.ID] = c
		for _, req := range c.Requirements {
			m.requireIndex[req.ID] = append(m.requireIndex[req.ID], c.ID)
		}
	}
	return m, nil
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
func Register(name string, path string) int {
	m, err := NewManager(path)
	if err != nil {
		return -1
	}

	mu.Lock()
	defer mu.Unlock()
	registry[name] = m
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
