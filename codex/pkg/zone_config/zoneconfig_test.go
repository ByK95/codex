package zoneconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestFile(t *testing.T, content string) string {
	tmpFile := "test_zones.json"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	assert.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(tmpFile) })
	return tmpFile
}

func TestZoneConfigManager_LoadAndGetters(t *testing.T) {
	jsonContent := `
[
	{
		"zone_id": "zone_1",
		"npcs": ["npc1", "npc2", "npc3"],
		"spawn_chance": 0.5,
		"max_count": 10
	},
	{
		"zone_id": "zone_2",
		"npcs": ["npcA"],
		"spawn_chance": 0.2,
		"max_count": 5
	}
]
`
	tmpFile := createTestFile(t, jsonContent)

	manager := GetManager(tmpFile)

	// Test MaxCount
	assert.Equal(t, int32(10), manager.GetMaxNPC("zone_1"))
	assert.Equal(t, int32(5), manager.GetMaxNPC("zone_2"))
	assert.Equal(t, int32(0), manager.GetMaxNPC("zone_missing"))

	// Test SpawnChance
	assert.Equal(t, float32(0.5), manager.GetSpawnChance("zone_1"))
	assert.Equal(t, float32(0.2), manager.GetSpawnChance("zone_2"))
	assert.Equal(t, float32(0), manager.GetSpawnChance("zone_missing"))

	// Test GetRandomNPCType
	npc := manager.GetRandomNPCType("zone_1")
	assert.Contains(t, []string{"npc1", "npc2", "npc3"}, npc)

	npc = manager.GetRandomNPCType("zone_2")
	assert.Equal(t, "npcA", npc)

	// Missing zone
	npc = manager.GetRandomNPCType("zone_missing")
	assert.Equal(t, "", npc)
}

func TestZoneConfigManager_LoadInvalidFile(t *testing.T) {
	tmpFile := createTestFile(t, `{ invalid json }`)
	manager := &ZoneConfigManager{path: tmpFile}
	err := manager.Load()
	assert.Error(t, err)
}
