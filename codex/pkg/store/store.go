package store

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

type ValueType int

const (
	IntType ValueType = iota
	FloatType
	BoolType
	StringType
)

type DrawItem struct {
	Name   string
	Chance int64
}

type StoreEntry struct {
	Type  ValueType   `json:"t"`
	Value interface{} `json:"v"`
}

type node struct {
	children map[string]*node
	entry    *StoreEntry // nil if not a leaf
}

var (
	globalStore *Store
	once   sync.Once
)

type Store struct {
	mu   sync.RWMutex
	root *node
	path string
	rng *rand.Rand
}

// GetStore returns the singleton Store instance, creating it if needed
func GetStore() *Store {
	once.Do(func() {
		globalStore = NewStore("./")
	})
	return globalStore
}

func NewStore(path string) *Store {
	return &Store{
		root: &node{children: make(map[string]*node)},
		path: path,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *Store) getOrCreateNode(key string) *node {
	cur := s.root
	parts := strings.Split(key, ".")
	for _, p := range parts {
		if cur.children[p] == nil {
			cur.children[p] = &node{children: make(map[string]*node)}
		}
		cur = cur.children[p]
	}
	return cur
}

func (s *Store) getNode(key string) *node {
	cur := s.root
	parts := strings.Split(key, ".")
	for _, p := range parts {
		next := cur.children[p]
		if next == nil {
			return nil
		}
		cur = next
	}
	return cur
}

// ---- Setters ----
func (s *Store) SetInt(key string, val int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := s.getOrCreateNode(key)
	n.entry = &StoreEntry{Type: IntType, Value: val}
}

func (s *Store) SetFloat(key string, val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := s.getOrCreateNode(key)
	n.entry = &StoreEntry{Type: FloatType, Value: math.Round(val*100) / 100}
}

func (s *Store) SetBool(key string, val bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := s.getOrCreateNode(key)
	n.entry = &StoreEntry{Type: BoolType, Value: val}
}

func (s *Store) SetString(key, val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := s.getOrCreateNode(key)
	n.entry = &StoreEntry{Type: StringType, Value: val}
}

// ---- Getters ----
func (s *Store) GetInt(key string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := s.getNode(key)
	if  n != nil && n.entry != nil && n.entry.Type == IntType {
		return n.entry.Value.(int64)
	}
	return 0
}

func (s *Store) GetFloat(key string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if n := s.getNode(key); n != nil && n.entry != nil && n.entry.Type == FloatType {
		return n.entry.Value.(float64)
	}
	return 0
}

func (s *Store) GetBool(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if n := s.getNode(key); n != nil && n.entry != nil && n.entry.Type == BoolType {
		return n.entry.Value.(bool)
	}
	return false
}

func (s *Store) GetString(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if n := s.getNode(key); n != nil && n.entry != nil && n.entry.Type == StringType {
		return n.entry.Value.(string)
	}
	return ""
}

func (s *Store) AddInt(key string, val int64) {
	current := s.GetInt(key)
	s.SetInt(key, current+val)
}

func (s *Store) SubInt(key string, val int64) bool {
	current := s.GetInt(key)
	if current < val {
		return false
	}
	s.SetInt(key, current-val)
	return true
}

func (s *Store) AddFloat(key string, val float64) {
	current := s.GetFloat(key)
	s.SetFloat(key, current+val)
}

func (s *Store) SubFloat(key string, val float64) bool {
	current := s.GetFloat(key)
	if current < val {
		return false
	}
	s.SetFloat(key, current-val)
	return true
}


// Keys returns all direct children keys under the given prefix.
// If prefix is empty, it returns top-level keys.
func (s *Store) Keys(prefix string) []string {
    s.mu.RLock()
    defer s.mu.RUnlock()

    cur := s.root
    if prefix != "" {
        parts := strings.Split(prefix, ".")
        for _, p := range parts {
            next := cur.children[p]
            if next == nil {
                return nil
            }
            cur = next
        }
    }

    keys := make([]string, 0, len(cur.children))
    for k := range cur.children {
        keys = append(keys, k)
    }
    return keys
}


func (s *Store) FullKeys(prefix string) []string {
    s.mu.RLock()
    defer s.mu.RUnlock()

    cur := s.root
    if prefix != "" {
        parts := strings.Split(prefix, ".")
        for _, p := range parts {
            next := cur.children[p]
            if next == nil {
                return nil
            }
            cur = next
        }
    }
	
    keys := make([]string, 0, len(cur.children))
    for k := range cur.children {
		if prefix != "" {
			s := fmt.Sprintf("%s.%s", prefix, k)
        	keys = append(keys, s)
		}else{
			keys = append(keys, k)
		}
    }
    return keys
}

func (s *Store) RandomSelect(prefix string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	children := s.FullKeys(prefix)

	var items []DrawItem
	var total int64

	for _, child := range children {
		name := s.GetString(child + ".name")
		chance := s.GetInt(child + ".chance")
		if name != "" && chance > 0 {
			items = append(items, DrawItem{Name: name, Chance: chance})
			total += chance
		}
	}

	if total == 0 || len(items) == 0 {
		return ""
	}

	roll := s.rng.Int63n(total)

	var cumulative int64
	for _, item := range items {
		cumulative += item.Chance
		if roll < cumulative {
			return item.Name
		}
	}

	return ""
}


// ---- Flatten/Unflatten ----
func flatten(prefix string, n *node, out map[string]StoreEntry) {
	if n.entry != nil {
		out[prefix] = *n.entry
	}
	for k, child := range n.children {
		var p string
		if prefix == "" {
			p = k
		} else {
			p = prefix + "." + k
		}
		flatten(p, child, out)
	}
}

func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	flat := make(map[string]StoreEntry)
	flatten("", s.root, flat)

	data, err := json.MarshalIndent(flat, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	raw := make(map[string]StoreEntry)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// reset root
	s.root = &node{children: make(map[string]*node)}

	// rebuild trie
	for k, e := range raw {
		n := s.getOrCreateNode(k)

		switch e.Type {
		case IntType:
			if f, ok := e.Value.(float64); ok {
				e.Value = int64(f)
			}
		case FloatType:
			if f, ok := e.Value.(float64); ok {
				e.Value = math.Round(f*100) / 100
			}
		case BoolType:
			if b, ok := e.Value.(bool); ok {
				e.Value = b
			} else {
				e.Value = false
			}
		case StringType:
			if str, ok := e.Value.(string); ok {
				e.Value = str
			} else {
				e.Value = ""
			}
		}

		// Assign the processed entry back to the node
		n.entry = &StoreEntry{
			Type:  e.Type,
			Value: e.Value,
		}
	}


	return nil
}

func (s *Store) LoadFromText(text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := make(map[string]StoreEntry)
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return err
	}

	// reset root
	s.root = &node{children: make(map[string]*node)}

	// rebuild trie
	for k, e := range raw {
		n := s.getOrCreateNode(k)

		switch e.Type {
		case IntType:
			if f, ok := e.Value.(float64); ok {
				e.Value = int64(f)
			}
		case FloatType:
			if f, ok := e.Value.(float64); ok {
				e.Value = math.Round(f*100) / 100
			}
		case BoolType:
			if b, ok := e.Value.(bool); ok {
				e.Value = b
			} else {
				e.Value = false
			}
		case StringType:
			if str, ok := e.Value.(string); ok {
				e.Value = str
			} else {
				e.Value = ""
			}
		}

		// assign processed entry
		n.entry = &StoreEntry{
			Type:  e.Type,
			Value: e.Value,
		}
	}

	return nil
}

