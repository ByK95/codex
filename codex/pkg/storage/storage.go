package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Handler defines the interface for package-specific data handling
type Handler interface {
    // Load is called with raw JSON data when loading from file
    Load(data json.RawMessage) error
    
    // Save should return the data to be saved as JSON
    Save() (any, error)
}

// HandlerFunc is a function type that implements Handler interface
type HandlerFunc struct {
    LoadFunc func(data json.RawMessage) error
    SaveFunc func() (any, error)
}

func (h HandlerFunc) Load(data json.RawMessage) error {
    if h.LoadFunc != nil {
        return h.LoadFunc(data)
    }
    return nil
}

func (h HandlerFunc) Save() (any, error) {
    if h.SaveFunc != nil {
        return h.SaveFunc()
    }
    return nil, nil
}

// StorageManager manages persistent data for multiple packages in a single file
type StorageManager struct {
    mu       sync.RWMutex
    handlers map[string]Handler
    filename string
    autoSave bool
    dirty    bool
}

var (
	manager *StorageManager
	once   sync.Once
)

// NewStorageManager creates a new storage manager instance
func SetStorageManagerPath(filename string) error {
    SM().filename = filename
	return SM().LoadAll()
}

// GetStore returns the singleton Store instance, creating it if needed
func SM() *StorageManager {
	once.Do(func() {
		manager = &StorageManager{
        handlers: make(map[string]Handler),
        filename: "",
        autoSave: false,
        dirty:    false,
    }
	})
	return manager
}

// Bind registers a handler for a specific key
func (s *StorageManager) Bind(key string, handler Handler) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.handlers[key] = handler
}

// BindFuncs registers load and save functions for a specific key
func (s *StorageManager) BindFuncs(key string, loadFunc func(json.RawMessage) error, saveFunc func() (any, error)) {
    s.Bind(key, HandlerFunc{
        LoadFunc: loadFunc,
        SaveFunc: saveFunc,
    })
}

// LoadAll reads the JSON file and calls load handlers for each key
func (s *StorageManager) LoadAll() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.loadAllUnsafe()
}

func (s *StorageManager) loadAllUnsafe() error {
    file, err := os.ReadFile(s.filename)
    if err != nil {
        if os.IsNotExist(err) {
            return nil // file doesn't exist yet
        }
        return fmt.Errorf("failed to read %s: %w", s.filename, err)
    }

    if len(file) == 0 {
        return nil // empty file
    }

    raw := make(map[string]json.RawMessage)
    if err := json.Unmarshal(file, &raw); err != nil {
        return fmt.Errorf("failed to unmarshal %s: %w", s.filename, err)
    }

    // Call load handlers for each key
    for key, handler := range s.handlers {
        if rawData, exists := raw[key]; exists {
            if err := handler.Load(rawData); err != nil {
                fmt.Printf("Warning: failed to load key %s: %v\n", key, err)
            }
        }
    }
    
    s.dirty = false
    return nil
}

// SaveAll calls save handlers and writes all data to the JSON file
func (s *StorageManager) SaveAll() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.saveAllUnsafe()
}

func (s *StorageManager) saveAllUnsafe() error {
    // Ensure directory exists
    if err := os.MkdirAll(filepath.Dir(s.filename), 0755); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }

    out := make(map[string]any)
    
    // Call save handlers for each key
    for key, handler := range s.handlers {
        data, err := handler.Save()
        if err != nil {
            return fmt.Errorf("failed to save key %s: %w", key, err)
        }
        if data != nil {
            out[key] = data
        }
    }

    buf, err := json.MarshalIndent(out, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal data: %w", err)
    }

    // Write to temporary file first, then rename (atomic operation)
    tempFile := s.filename + ".tmp"
    if err := os.WriteFile(tempFile, buf, 0644); err != nil {
        return fmt.Errorf("failed to write temp file: %w", err)
    }

    if err := os.Rename(tempFile, s.filename); err != nil {
        os.Remove(tempFile) // cleanup
        return fmt.Errorf("failed to rename temp file: %w", err)
    }

    s.dirty = false
    return nil
}