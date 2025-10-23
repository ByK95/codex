package store

import (
	"codex/pkg/iterator"
	"codex/pkg/storage"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type ValueType int
var iter *iterator.Iterator[string]

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

type GroupedStore struct {
    Strings map[string]string  `json:"strings"`
    Ints    map[string]int64   `json:"ints"`
    Floats  map[string]float64 `json:"floats"`
    Bools   map[string]bool    `json:"bools"`
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
	rng *rand.Rand
}

// GetStore returns the singleton Store instance, creating it if needed
func GetStore() *Store {
	once.Do(func() {
		globalStore = NewStore()
	})
	return globalStore
}

func init() {
	st := GetStore()
	storage.SM().BindFuncs(
		"store",
		st.Load,
		func() (any, error) {
			return st.Save()
		},
	)
}

func NewStore() *Store {
	return &Store{
		root: &node{children: make(map[string]*node)},
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

func (s *Store) ReleaseBool(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if n := s.getNode(key); n != nil && n.entry != nil && n.entry.Type == BoolType{
		if n.entry.Value.(bool) {
			n.entry.Value = false
			return true
		}
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

// Clear removes all keys (and their subkeys) under the given prefix
func (s *Store) Clear(prefix string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if prefix == "" {
		// clear everything
		s.root = &node{children: make(map[string]*node)}
		return
	}

	parts := strings.Split(prefix, ".")
	cur := s.root
	for i := 0; i < len(parts)-1; i++ {
		next := cur.children[parts[i]]
		if next == nil {
			return // prefix doesn't exist
		}
		cur = next
	}

	// delete the last part
	delete(cur.children, parts[len(parts)-1])
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

func (s *Store) Save() ([]byte, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    grouped := GroupedStore{
        Strings: make(map[string]string),
        Ints:    make(map[string]int64),
        Floats:  make(map[string]float64),
        Bools:   make(map[string]bool),
    }

    // traverse trie and group by type
    s.traverseAndGroup(s.root, "", &grouped)

    return json.MarshalIndent(grouped, "", "  ")
}

func (s *Store) traverseAndGroup(n *node, prefix string, grouped *GroupedStore) {
    if n.entry != nil {
        switch n.entry.Type {
        case StringType:
            if v, ok := n.entry.Value.(string); ok {
                grouped.Strings[prefix] = v
            }
        case IntType:
            if v, ok := n.entry.Value.(int64); ok {
                grouped.Ints[prefix] = v
            }
        case FloatType:
            if v, ok := n.entry.Value.(float64); ok {
                grouped.Floats[prefix] = v
            }
        case BoolType:
            if v, ok := n.entry.Value.(bool); ok {
                grouped.Bools[prefix] = v
            }
        }
    }

    for key, child := range n.children {
        childPrefix := prefix
        if childPrefix != "" {
            childPrefix += "."
        }
        childPrefix += key
        s.traverseAndGroup(child, childPrefix, grouped)
    }
}

func (s *Store) Load(data json.RawMessage) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    grouped := GroupedStore{
        Strings: make(map[string]string),
        Ints:    make(map[string]int64),
        Floats:  make(map[string]float64),
        Bools:   make(map[string]bool),
    }

    if err := json.Unmarshal(data, &grouped); err != nil {
        return err
    }

    // reset root
    s.root = &node{children: make(map[string]*node)}

    // rebuild trie from strings
    for k, v := range grouped.Strings {
        n := s.getOrCreateNode(k)
        n.entry = &StoreEntry{
            Type:  StringType,
            Value: v,
        }
    }

    // rebuild trie from ints
    for k, v := range grouped.Ints {
        n := s.getOrCreateNode(k)
        n.entry = &StoreEntry{
            Type:  IntType,
            Value: v,
        }
    }

    // rebuild trie from floats
    for k, v := range grouped.Floats {
        n := s.getOrCreateNode(k)
        n.entry = &StoreEntry{
            Type:  FloatType,
            Value: math.Round(v*100) / 100,
        }
    }

    // rebuild trie from bools
    for k, v := range grouped.Bools {
        n := s.getOrCreateNode(k)
        n.entry = &StoreEntry{
            Type:  BoolType,
            Value: v,
        }
    }

    return nil
}

func (s *Store) LoadFromText(text string) error {
    return s.Load(json.RawMessage(text))
}

func InitGetFullKeysIter(prefix string){
	keys := GetStore().FullKeys(prefix)
	iter = iterator.NewIterator(keys)
}

func Next() string{
	if iter == nil {
		return ""
	}
	val, _ := iter.Next()
	return val
}