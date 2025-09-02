package zoneconfig

import (
	"testing"
	"encoding/json"
	"github.com/stretchr/testify/assert"
)

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
	Load(json.RawMessage(jsonContent))
	manager := GetManager()

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


func TestZoneConfigManager_LoadZone(t *testing.T) {
	jsonContent := `
	[
		{
			"zone_id": "1",
			"npcs": ["meteor"],
			"spawn_chance": 0.0017,
			"max_count": 10
		}
	]
	`
	manager := GetManager()
	Load(json.RawMessage(jsonContent))

	assert.Equal(t, int32(10), manager.GetMaxNPC("1"))
}

func TestZoneConfigManager_LoadInvalidFile(t *testing.T) {
	err := Load(json.RawMessage(`{ invalid json }`))
	assert.NotNil(t, err)
}
