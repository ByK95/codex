package store

import (
	"encoding/json"
	"math"
	"os"
	"sync"
)

type ValueType int
var store *Store
var once sync.Once

// GetManager returns the singleton instance
func GetStore() *Store {
	once.Do(func() {
		store = NewStore("./")
	})
	return store
}

const (
	IntType ValueType = iota
	FloatType
	BoolType
	StringType
)

type StoreEntry struct {
	Type  ValueType   `json:"t"`
	Value interface{} `json:"v"`
}

type Store struct {
	mu      sync.RWMutex
	objects map[string]StoreEntry
	path    string
}

func NewStore(path string) *Store {
	return &Store{
		objects: make(map[string]StoreEntry),
		path:    path,
	}
}

// ---- Setters ----
func (s *Store) SetInt(key string, val int64) {
	s.mu.Lock()
	s.objects[key] = StoreEntry{Type: IntType, Value: val}
	s.mu.Unlock()
}

func (s *Store) SetFloat(key string, val float64) {
	s.mu.Lock()
	val = math.Round(val*100) / 100
	s.objects[key] = StoreEntry{Type: FloatType, Value: val}
	s.mu.Unlock()
}

func (s *Store) SetBool(key string, val bool) {
	s.mu.Lock()
	s.objects[key] = StoreEntry{Type: BoolType, Value: val}
	s.mu.Unlock()
}

func (s *Store) SetString(key, val string) {
	s.mu.Lock()
	s.objects[key] = StoreEntry{Type: StringType, Value: val}
	s.mu.Unlock()
}

// ---- Getters ----
func (s *Store) GetInt(key string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.objects[key]; ok && e.Type == IntType {
		return e.Value.(int64)
	}
	return 0
}

func (s *Store) AddInt(key string, val int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.objects[key]
	if !ok || e.Type != IntType {
		e = StoreEntry{Type: IntType, Value: int64(0)}
	}
	e.Value = e.Value.(int64) + val
	s.objects[key] = e
}

func (s *Store) SubInt(key string, val int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.objects[key]
	if !ok || e.Type != IntType {
		return false
	}
	cur := e.Value.(int64)
	if cur < val {
		return false
	}
	e.Value = cur - val
	s.objects[key] = e
	return true
}

func (s *Store) GetFloat(key string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.objects[key]; ok && e.Type == FloatType {
		return e.Value.(float64)
	}
	return 0
}

func (s *Store) AddFloat(key string, val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.objects[key]
	if !ok || e.Type != FloatType {
		e = StoreEntry{Type: FloatType, Value: float64(0)}
	}
	e.Value = math.Round((e.Value.(float64)+val)*100) / 100
	s.objects[key] = e
}

func (s *Store) SubFloat(key string, val float64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.objects[key]
	if !ok || e.Type != FloatType {
		return false
	}
	cur := e.Value.(float64)
	if cur < val {
		return false
	}
	e.Value = math.Round((cur-val)*100) / 100
	s.objects[key] = e
	return true
}

func (s *Store) GetBool(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.objects[key]; ok && e.Type == BoolType {
		return e.Value.(bool)
	}
	return false
}

func (s *Store) GetString(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.objects[key]; ok && e.Type == StringType {
		return e.Value.(string)
	}
	return ""
}

// ---- Persistence ----
func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.objects, "", "  ")
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

	for k, e := range raw {
		switch e.Type {
		case IntType:
			if f, ok := e.Value.(float64); ok { // JSON numbers -> float64
				e.Value = int64(f)
			}
		case FloatType:
			if f, ok := e.Value.(float64); ok {
				e.Value = math.Round(f*100) / 100
			}
		}
		s.objects[k] = e
	}

	return nil
}
