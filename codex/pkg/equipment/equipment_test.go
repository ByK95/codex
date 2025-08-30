package equipment

import (
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEquipmentManager(t *testing.T) {
	em := NewEquipmentManager()
	assert.NotNil(t, em)
	assert.NotNil(t, em.slots)
	assert.Equal(t, 0, len(em.slots))
}

func TestGetManager_Singleton(t *testing.T) {
	globalManager = nil
	once = sync.Once{}

	em1 := GetManager()
	em2 := GetManager()

	assert.NotNil(t, em1)
	assert.Same(t, em1, em2)
}

func TestDefineSlot(t *testing.T) {
	em := NewEquipmentManager()

	assert.True(t, em.DefineSlot("head", 2))
	assert.False(t, em.DefineSlot("body", 0))
	assert.False(t, em.DefineSlot("legs", -1))

	assert.True(t, em.EquipItem("head", "helmet1"))
	assert.True(t, em.EquipItem("head", "helmet2"))
	assert.True(t, em.DefineSlot("head", 1))

	items := em.GetEquippedItems("head")
	assert.Equal(t, 1, len(items))
}

func TestRemoveSlotDefinition(t *testing.T) {
	em := NewEquipmentManager()

	assert.False(t, em.RemoveSlotDefinition("head"))

	assert.True(t, em.DefineSlot("head", 2))
	assert.True(t, em.EquipItem("head", "helmet1"))
	assert.True(t, em.RemoveSlotDefinition("head"))
}

func TestEquipItem(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot("head", 2)

	assert.True(t, em.EquipItem("head", "helmet1"))
	items := em.GetEquippedItems("head")
	assert.Contains(t, items, "helmet1")

	assert.False(t, em.EquipItem("head", ""))
	assert.False(t, em.EquipItem("nonexistent", "shield1"))
	assert.False(t, em.EquipItem("head", "helmet1"))

	assert.True(t, em.EquipItem("head", "helmet2"))
	assert.False(t, em.EquipItem("head", "helmet3"))
}

func TestUnequipItem(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot("head", 3)
	em.EquipItem("head", "helmet1")
	em.EquipItem("head", "helmet2")
	em.EquipItem("head", "helmet3")

	assert.True(t, em.UnequipItem("head", "helmet2"))
	items := em.GetEquippedItems("head")
	assert.NotContains(t, items, "helmet2")
	assert.Contains(t, items, "helmet1")
	assert.Contains(t, items, "helmet3")

	assert.False(t, em.UnequipItem("head", "nonexistent"))
	assert.False(t, em.UnequipItem("nonexistent", "helmet1"))
}

func TestGetEquippedItems(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot("head", 3)

	items := em.GetEquippedItems("head")
	assert.Empty(t, items)

	items = em.GetEquippedItems("nonexistent")
	assert.Empty(t, items)

	em.EquipItem("head", "helmet1")
	em.EquipItem("head", "helmet2")
	items = em.GetEquippedItems("head")
	assert.Len(t, items, 2)
	assert.Contains(t, items, "helmet1")
	assert.Contains(t, items, "helmet2")

	items[0] = "modified"
	original := em.GetEquippedItems("head")
	assert.NotContains(t, original, "modified")
}

func TestIsSlotEmpty(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot("head", 2)

	assert.True(t, em.IsSlotEmpty("head"))
	assert.True(t, em.IsSlotEmpty("nonexistent"))

	em.EquipItem("head", "helmet1")
	assert.False(t, em.IsSlotEmpty("head"))
}

func TestIsSlotFull(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot("head", 2)

	assert.False(t, em.IsSlotFull("head"))

	em.EquipItem("head", "helmet1")
	assert.False(t, em.IsSlotFull("head"))

	em.EquipItem("head", "helmet2")
	assert.True(t, em.IsSlotFull("head"))

	assert.False(t, em.IsSlotFull("nonexistent"))
}

func TestIsItemEquipped(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot("head", 2)
	em.EquipItem("head", "helmet1")

	assert.True(t, em.IsItemEquipped("head", "helmet1"))
	assert.False(t, em.IsItemEquipped("head", "helmet2"))
	assert.False(t, em.IsItemEquipped("nonexistent", "helmet1"))
}

func TestGetAllEquippedItems(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot("head", 2)
	em.DefineSlot("body", 1)
	em.DefineSlot("legs", 3)

	items := em.GetAllEquippedItems()
	assert.Empty(t, items)

	em.EquipItem("head", "helmet1")
	em.EquipItem("head", "helmet2")
	em.EquipItem("body", "armor1")
	em.EquipItem("legs", "boots1")
	em.EquipItem("legs", "boots2")

	items = em.GetAllEquippedItems()
	expected := []string{"helmet1", "helmet2", "armor1", "boots1", "boots2"}
	sort.Strings(items)
	sort.Strings(expected)
	assert.Equal(t, expected, items)
}

func TestGetAllSlotTypes(t *testing.T) {
	em := NewEquipmentManager()
	types := em.GetAllSlotTypes()
	assert.Empty(t, types)

	em.DefineSlot("head", 2)
	em.DefineSlot("body", 1)
	em.DefineSlot("legs", 3)

	types = em.GetAllSlotTypes()
	expected := []string{"head", "body", "legs"}
	sort.Strings(types)
	sort.Strings(expected)
	assert.Equal(t, expected, types)
}

func TestClearAndClearSlot(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot("head", 2)
	em.DefineSlot("body", 1)
	em.EquipItem("head", "helmet1")
	em.EquipItem("head", "helmet2")
	em.EquipItem("body", "armor1")

	em.Clear()
	assert.True(t, em.IsSlotEmpty("head"))
	assert.True(t, em.IsSlotEmpty("body"))
	assert.Empty(t, em.GetAllEquippedItems())

	em.EquipItem("head", "helmet1")
	em.EquipItem("body", "armor1")
	assert.True(t, em.ClearSlot("head"))
	assert.True(t, em.IsSlotEmpty("head"))
	assert.False(t, em.IsSlotEmpty("body"))
	assert.False(t, em.ClearSlot("nonexistent"))
}

func TestReset(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot("head", 2)
	em.DefineSlot("body", 1)
	em.EquipItem("head", "helmet1")
	em.EquipItem("body", "armor1")

	em.Reset()
	assert.Empty(t, em.GetAllSlotTypes())
	assert.Empty(t, em.GetAllEquippedItems())
}

func TestGetSlotAvailability(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot("head", 3)

	availability := em.GetSlotAvailability("head")
	assert.Equal(t, 3, availability)

	em.EquipItem("head", "helmet1")
	availability = em.GetSlotAvailability("head")
	assert.Equal(t, 2, availability)

	em.EquipItem("head", "helmet2")
	em.EquipItem("head", "helmet3")
	availability = em.GetSlotAvailability("head")
	assert.Equal(t, 0, availability)

	assert.Equal(t, 0, em.GetSlotAvailability("nonexistent"))
}

func TestEquipmentManagerIterator(t *testing.T) {
	em := &EquipmentManager{
		mu:    sync.RWMutex{},
		slots: make(map[string]*SlotConfig),
	}
	em.slots["head"] = &SlotConfig{ItemIDS: []string{"helmet1", "helmet2"}, MaxSlots: 2}
	em.slots["body"] = &SlotConfig{ItemIDS: []string{"armor1"}, MaxSlots: 2}

	em.ResetIterator()

	expected := []string{"armor1", "helmet1", "helmet2"}
	var got []string

	for {
		id := em.NextEquippedItem()
		if id == "" {
			break
		}
		got = append(got, id)
	}

	sort.Strings(expected)
	sort.Strings(got)
	assert.Equal(t, expected, got)
	assert.Equal(t, "", em.NextEquippedItem())

	em.ResetIterator()
	assert.Equal(t, "armor1", em.NextEquippedItem())
}