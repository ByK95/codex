package store

import (
	"os"
	"testing"
	"fmt"
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
	s.SetInt("player.progress.level", 7) // nested key

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
	if v := s2.GetInt("player.progress.level"); v != 7 {
		t.Errorf("expected 7, got %d", v)
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

func TestKeys(t *testing.T) {
	s := NewStore("")

	// Setup nested structure
	s.SetInt("player.progress.level", 7)
	s.SetString("player.name", "Bayram")
	s.SetFloat("player.speed", 1.5)
	s.SetInt("currency.gold", 200)

	// Test top-level
	top := s.Keys("")
	expectedTop := map[string]bool{"player": true, "currency": true}
	if len(top) != len(expectedTop) {
		t.Errorf("expected %d top-level keys, got %d", len(expectedTop), len(top))
	}
	for _, k := range top {
		if !expectedTop[k] {
			t.Errorf("unexpected top-level key: %s", k)
		}
	}

	// Test child keys under "player"
	playerChildren := s.Keys("player")
	expectedPlayer := map[string]bool{"progress": true, "name": true, "speed": true}
	if len(playerChildren) != len(expectedPlayer) {
		t.Errorf("expected %d player children, got %d", len(expectedPlayer), len(playerChildren))
	}
	for _, k := range playerChildren {
		if !expectedPlayer[k] {
			t.Errorf("unexpected player child key: %s", k)
		}
	}

	// Test nested child under "player.progress"
	progressChildren := s.Keys("player.progress")
	expectedProgress := map[string]bool{"level": true}
	if len(progressChildren) != len(expectedProgress) {
		t.Errorf("expected %d progress children, got %d", len(expectedProgress), len(progressChildren))
	}
	if progressChildren[0] != "level" {
		t.Errorf("expected 'level', got %s", progressChildren[0])
	}

	// Nonexistent prefix should return empty list
	none := s.Keys("does.not.exist")
	if len(none) != 0 {
		t.Errorf("expected no keys, got %v", none)
	}
}

func TestStoreFullKeys(t *testing.T) {
    s := NewStore("test.json")

    // Insert some values
    s.SetString("player.name", "TestPlayer")
    s.SetInt("player.level", 7)
    s.SetInt("currency.gold", 200)

    tests := []struct {
        prefix   string
        expected []string
    }{
        {
            prefix:   "",
            expected: []string{"player", "currency"},
        },
        {
            prefix:   "player",
            expected: []string{"player.name", "player.level"},
        },
        {
            prefix:   "currency",
            expected: []string{"currency.gold"},
        },
        {
            prefix:   "quest", // doesn't exist
            expected: nil,
        },
    }

    for _, tt := range tests {
        got := s.FullKeys(tt.prefix)
        if !equalUnordered(got, tt.expected) {
            t.Errorf("FullKeys(%q) = %v, want %v", tt.prefix, got, tt.expected)
        }
    }
}

// helper to compare slices ignoring order
func equalUnordered(a, b []string) bool {
    if len(a) != len(b) {
        return false
    }
    m := make(map[string]int)
    for _, v := range a {
        m[v]++
    }
    for _, v := range b {
        if m[v] == 0 {
            return false
        }
        m[v]--
    }
    return true
}

func TestLoadFromText(t *testing.T) {
	text := `{
		"currency.gold": {"t": 0, "v": 200},
		"player.name": {"t": 3, "v": "TestPlayer"},
		"player.progress.level": {"t": 0, "v": 7},
		"player.speed": {"t": 1, "v": 1.23},
		"quest.completed": {"t": 2, "v": false}
	}`

	s := NewStore(".")
	err := s.LoadFromText(text)
	if err != nil {
		t.Fatalf("LoadFromText failed: %v", err)
	}

	// integers
	if v := s.GetInt("currency.gold"); v != 200 {
		t.Errorf("expected 200, got %v ", v)
	}
	if v := s.GetInt("player.progress.level"); v != 7 {
		t.Errorf("expected 7, got %v ", v)
	}

	// float
	if v := s.GetFloat("player.speed"); v != 1.23 {
		t.Errorf("expected 1.23, got %v ", v)
	}

	// string
	if v := s.GetString("player.name");  v != "TestPlayer" {
		t.Errorf("expected 'TestPlayer', got %q ", v)
	}

	// bool
	if v := s.GetBool("quest.completed"); v != false{
		t.Errorf("expected false, got %v", v)
	}
}


func TestRandomSelection(t *testing.T) {
	text := `{
		"galactic_draw.1.name": {"t": 3, "v": "ship.starlance"},
		"galactic_draw.1.chance": {"t": 0, "v": 90},
		"galactic_draw.2.name": {"t": 3, "v": "weapon.laser"},
		"galactic_draw.2.chance": {"t": 0, "v": 90},
		"galactic_draw.3.name": {"t": 3, "v": "extras.radar"},
		"galactic_draw.3.chance": {"t": 0, "v": 90},
		"galactic_draw.4.name": {"t": 3, "v": "captain.skywalker"},
		"galactic_draw.4.chance": {"t": 0, "v": 90},
		"galactic_draw.5.name": {"t": 3, "v": "captain.luke"},
		"galactic_draw.5.chance": {"t": 0, "v": 90},
		"galactic_draw.6.name": {"t": 3, "v": "weapon.rocket_launcher"},
		"galactic_draw.6.chance": {"t": 0, "v": 10},
		"galactic_draw.7.name": {"t": 3, "v": "weapon.ion_ball"},
		"galactic_draw.7.chance": {"t": 0, "v": 1}
	}`

	s := NewStore(".")
	err := s.LoadFromText(text)

	if err != nil {
		t.Fatalf("LoadFromText failed: %v", err)
	}

	const runs = 200
	counts := make(map[string]int)

	for i := 0; i < runs; i++ {
		res := s.RandomSelect("galactic_draw")
		if res == "" {
			t.Fatalf("RandomSelect returned empty at iteration %d", i)
		}
		counts[res]++
	}

	// Assert only expected keys are drawn
	expected := map[string]bool{
		"ship.starlance":       true,
		"weapon.laser":         true,
		"extras.radar":         true,
		"captain.skywalker":    true,
		"captain.luke":         true,
		"weapon.rocket_launcher": true,
		"weapon.ion_ball":      true,
	}
	for key := range counts {
		if !expected[key] {
			t.Errorf("Unexpected draw result: %q", key)
		}
	}

	// Print result distribution
	fmt.Println("Distribution after", runs, "runs:")
	for k, v := range counts {
		fmt.Printf("%-25s %d\n", k, v)
	}
}