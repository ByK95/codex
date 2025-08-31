package helpers

import (
	"fmt"
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
	"codex/pkg/crafting"
	"codex/pkg/equipment"
	"codex/pkg/store"
)


func writeTestFile(t *testing.T, content string) string {
	tmpfile, err := os.CreateTemp("", "crafting_test_*.json")
	assert.NoError(t, err)

	_, err = tmpfile.Write([]byte(content))
	assert.NoError(t, err)

	err = tmpfile.Close()
	assert.NoError(t, err)

	return tmpfile.Name()
}

func TestHelpers(t *testing.T) {
	var testJSON = `[
		{"id":"weapon.laser.2","requirements":[{"id":"weapon.laser","qty":1}]},
		{"id":"weapon.laser.3","requirements":[{"id":"weapon.laser.2","qty":1}]},
		{"id":"weapon.laser","requirements":[{"id":"","qty":0}]}
	]`
	path := writeTestFile(t, testJSON)
	defer os.Remove(path)
	s := store.GetStore()
	s.SetString("weapon.laser.slot_type", "weapon")
	s.SetString("weapon.laser.2.slot_type", "weapon")
	s.SetString("weapon.laser.3.slot_type", "weapon")
	

	// Register manager
	crafting.Register(upgrades, path)

	equipment.GetManager().DefineSlot("weapon", 1)

	upgrades := GetUpgrades()
	if len(upgrades) != 1 || upgrades[0] != "weapon.laser" {
		t.Errorf("Expected 1 upgrade 'weapon.laser', got: %v", upgrades)
	}

	ok := UpgrageItem(upgrades[0])
	if !ok {
		t.Errorf("Failed to upgrade weapon.laser")
	}
	
	upgrades = GetUpgrades()
	if len(upgrades) != 1 || upgrades[0] != "weapon.laser.2" {
		t.Errorf("Expected 1 upgrade 'weapon.laser', got: %v", upgrades)
	}

	ok = UpgrageItem(upgrades[0])
	assert.True(t, ok, "Failed to upgrade weapon.laser.2")
}

func TestHelperIterators(t *testing.T) {
	var testJSON = `[
		{"id":"weapon.laser.2","requirements":[{"id":"weapon.laser","qty":1}]},
		{"id":"weapon.laser.3","requirements":[{"id":"weapon.laser.2","qty":1}]},
		{"id":"weapon.laser","requirements":[{"id":"","qty":0}]},
		{"id":"captain.kaori","requirements":[{"id":"","qty":0}]},
		{"id":"extra.radio","requirements":[{"id":"","qty":0}]},
		{"id":"extra.gps","requirements":[{"id":"","qty":0}]},
		{"id":"weapon.rocket_launcher","requirements":[{"id":"","qty":0}]}
	]`
	path := writeTestFile(t, testJSON)
	defer os.Remove(path)
	s := store.GetStore()
	s.SetString("weapon.laser.slot_type", "weapon")
	s.SetString("weapon.rocket_launcher.slot_type", "weapon")
	s.SetString("weapon.laser.2.slot_type", "weapon")
	s.SetString("weapon.laser.3.slot_type", "weapon")
	s.SetString("captain.kaori.slot_type", "captain")
	s.SetString("extra.radio.slot_type", "extra")
	

	// Register manager
	crafting.Register(upgrades, path)

	equipment.GetManager().DefineSlot("weapon", 1)
	equipment.GetManager().DefineSlot("captain", 1)
	equipment.GetManager().DefineSlot("extra", 1)

	GetUpgradeSelections(3)

	selection1 := GetNextSelections()
	selection2 := GetNextSelections()
	selection3 := GetNextSelections()
	selection4 := GetNextSelections() // should be empty
	fmt.Println("Selections:", selection1, selection2, selection3, selection4)
}