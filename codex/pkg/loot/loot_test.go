package loot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRollLootGuaranteed(t *testing.T) {
	ResetPity()
	items := []LootRow{
		{ID: 1, Chance: 1.0, Pity: 0},
		{ID: 2, Chance: 1.0, Pity: 0},
	}

	results := RollLoot(items)

	assert.Greater(t, results[int32(1)], int32(0), "Guaranteed item 1 should drop")
	assert.Greater(t, results[int32(2)], int32(0), "Guaranteed item 2 should drop")
}

func TestRollLootPity(t *testing.T) {
	ResetPity()
	items := []LootRow{
		{ID: 1, Chance: 0.0, Pity: 3}, // zero chance, pity triggers
	}

	for i := 0; i < 3; i++ {
		results := RollLoot(items)
		assert.Empty(t, results, "No drop before pity threshold on roll %d", i+1)
	}

	results := RollLoot(items)
	assert.NotEmpty(t, results, "Drop expected due to pity")
	assert.Greater(t, results[int32(1)], int32(0), "Pity should cause drop for item 1")

	results = RollLoot(items)
	assert.Empty(t, results, "No drop after pity reset")
}

func TestRollLootChance(t *testing.T) {
	ResetPity()
	items := []LootRow{
		{ID: 1, Chance: 1.0, Pity: 3}, // guaranteed drop
	}

	results := RollLoot(items)
	assert.Greater(t, results[int32(1)], int32(0), "Guaranteed drop expected")

	assert.Equal(t, int32(0), pityCounts[1], "Pity count should reset after drop")
}

func TestResetPity(t *testing.T) {
	items := []LootRow{
		{ID: 1, Chance: 0.0, Pity: 1},
	}

	RollLoot(items)
	assert.Equal(t, int32(1), pityCounts[1], "Pity count should increment")

	ResetPity()
	assert.Empty(t, pityCounts, "Pity counts should be reset")
}
