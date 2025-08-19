package store

import (
	"os"
	"testing"

)

func TestStoreBasicOperations(t *testing.T) {
	s := NewStore("")

	// Int
	s.SetInt("currency.gold", 100)
	if v := s.GetInt("currency.gold"); v != 100 {
		t.Errorf("expected 100, got %d", v)
	}

	// Float
	s.SetFloat("player.speed", 3.14)
	if v := s.GetFloat("player.speed"); v != 3.14 {
		t.Errorf("expected 3.14, got %f", v)
	}

	// Bool
	s.SetBool("quest.completed", true)
	if v := s.GetBool("quest.completed"); v != true {
		t.Errorf("expected true, got %v", v)
	}

	// String
	s.SetString("player.name", "Bayram")
	if v := s.GetString("player.name"); v != "Bayram" {
		t.Errorf("expected 'Bayram', got %s", v)
	}
}

func TestStorePersistence(t *testing.T) {
	tmpFile := "test_json"
	defer os.Remove(tmpFile)

	s := NewStore(tmpFile)
	s.SetInt("currency.gold", 200)
	s.SetFloat("player.speed", 1.23)
	s.SetBool("quest.completed", false)
	s.SetString("player.name", "TestPlayer")

	if err := s.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Create new store and load
	s2 := NewStore(tmpFile)
	if err := s2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Check values
	if v := s2.GetInt("currency.gold"); v != 200 {
		t.Errorf("expected 200, got %d", v)
	}
	if v := s2.GetFloat("player.speed"); v != 1.23 {
		t.Errorf("expected 1.23, got %f", v)
	}
	if v := s2.GetBool("quest.completed"); v != false {
		t.Errorf("expected false, got %v", v)
	}
	if v := s2.GetString("player.name"); v != "TestPlayer" {
		t.Errorf("expected 'TestPlayer', got %s", v)
	}
}

func TestOverwriteValues(t *testing.T) {
	s := NewStore("")

	s.SetInt("val", 10)
	s.SetInt("val", 20)
	if v := s.GetInt("val"); v != 20 {
		t.Errorf("expected 20 after overwrite, got %d", v)
	}

	s.SetString("val2", "a")
	s.SetString("val2", "b")
	if v := s.GetString("val2"); v != "b" {
		t.Errorf("expected 'b' after overwrite, got %s", v)
	}
}

func TestTypeMismatchReturnsDefault(t *testing.T) {
	s := NewStore("")

	s.SetInt("val", 42)
	if v := s.GetFloat("val"); v != 0 {
		t.Errorf("expected 0 on type mismatch, got %f", v)
	}
	if v := s.GetBool("val"); v != false {
		t.Errorf("expected false on type mismatch, got %v", v)
	}
	if v := s.GetString("val"); v != "" {
		t.Errorf("expected empty string on type mismatch, got %s", v)
	}
}

func TestNonExistentKeysReturnDefault(t *testing.T) {
	s := NewStore("")

	if v := s.GetInt("does.not.exist"); v != 0 {
		t.Errorf("expected 0 for missing int key, got %d", v)
	}
	if v := s.GetFloat("does.not.exist"); v != 0 {
		t.Errorf("expected 0.0 for missing float key, got %f", v)
	}
	if v := s.GetBool("does.not.exist"); v != false {
		t.Errorf("expected false for missing bool key, got %v", v)
	}
	if v := s.GetString("does.not.exist"); v != "" {
		t.Errorf("expected empty string for missing string key, got %s", v)
	}
}

func TestSaveAndLoadEmptyStore(t *testing.T) {
	tmpFile := "empty_store.json"
	defer os.Remove(tmpFile)

	s := NewStore(tmpFile)
	if err := s.Save(); err != nil {
		t.Fatalf("Save failed on empty store: %v", err)
	}

	s2 := NewStore(tmpFile)
	if err := s2.Load(); err != nil {
		t.Fatalf("Load failed on empty store: %v", err)
	}

	// Nothing should exist
	if v := s2.GetInt("currency.gold"); v != 0 {
		t.Errorf("expected 0 for missing int after load, got %d", v)
	}
}

func TestPersistenceOverwrite(t *testing.T) {
	tmpFile := "overwrite_store.json"
	defer os.Remove(tmpFile)

	s := NewStore(tmpFile)
	s.SetInt("currency.gold", 50)
	_ = s.Save()

	// Overwrite
	s2 := NewStore(tmpFile)
	_ = s2.Load()
	s2.SetInt("currency.gold", 999)
	_ = s2.Save()

	s3 := NewStore(tmpFile)
	_ = s3.Load()
	if v := s3.GetInt("currency.gold"); v != 999 {
		t.Errorf("expected 999 after overwrite and reload, got %d", v)
	}
}

func TestMultipleTypesSameKey(t *testing.T) {
	s := NewStore("")

	s.SetInt("key", 100)
	s.SetString("key", "string-value")
	if v := s.GetString("key"); v != "string-value" {
		t.Errorf("expected 'string-value', got %s", v)
	}
	if v := s.GetInt("key"); v != 0 {
		t.Errorf("expected 0 after overwriting int with string, got %d", v)
	}
}

func TestFileNotFoundLoad(t *testing.T) {
	s := NewStore("non_existent_file.json")
	err := s.Load()
	if err == nil {
		t.Errorf("expected error when loading non-existent file, got nil")
	}
}

