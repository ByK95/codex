package crafting

import (
	"encoding/json"
	"testing"
	"github.com/stretchr/testify/assert"
)

var testJSON = `{
	"test": [
		{"id":"axe","requirements":[{"id":"wood","qty":3},{"id":"iron","qty":2}]},
		{"id":"pickaxe","requirements":[{"id":"wood","qty":2},{"id":"stone","qty":3}]},
		{"id":"potion","requirements":[{"id":"herb","qty":5}]}
	],
	"other": [
		{"id":"elixir","requirements":[{"id":"herb","qty":7}]}
	]
}`

func loadTestData(t *testing.T)  {
	LoadManagers(json.RawMessage(testJSON))
}

func TestManager_LoadAndLookup(t *testing.T) {
	loadTestData(t)

	// Retrieve manager
	m, ok := Get("test")
	assert.True(t, ok)
	assert.NotNil(t, m)

	// Forward lookup
	axe, found := m.GetCraftable("axe")
	assert.True(t, found)
	assert.Equal(t, 2, len(axe.Requirements))
	assert.Equal(t, "wood", axe.Requirements[0].ID)
	assert.Equal(t, 3, axe.Requirements[0].Qty)

	// Reverse lookup
	woodItems := m.FindByRequirement("wood")
	assert.Len(t, woodItems, 2) // axe + pickaxe
	var ids []string
	for _, item := range woodItems {
		ids = append(ids, item.ID)
	}
	assert.Contains(t, ids, "axe")
	assert.Contains(t, ids, "pickaxe")

	herbItems := m.FindByRequirement("herb")
	assert.Len(t, herbItems, 1)
	assert.Equal(t, "potion", herbItems[0].ID)

	// Non-existent item
	_, found = m.GetCraftable("nonexistent")
	assert.False(t, found)
	assert.Empty(t, m.FindByRequirement("nonexistent"))
}

func TestRegistry_ResetSingleAndAll(t *testing.T) {
	_ = Register("m1", &Manager{})
	_ = Register("m2", &Manager{})

	m1, ok := Get("m1")
	assert.True(t, ok)
	assert.NotNil(t, m1)

	m2, ok := Get("m2")
	assert.True(t, ok)
	assert.NotNil(t, m2)

	// Reset single
	Reset("m1")
	_, ok = Get("m1")
	assert.False(t, ok)

	_, ok = Get("m2")
	assert.True(t, ok) // m2 still exists

	// Reset all
	ResetAll()
	_, ok = Get("m2")
	assert.False(t, ok)
}
