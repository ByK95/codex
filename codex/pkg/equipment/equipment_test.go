package equipment

import (
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEquipmentManager(t *testing.T) {
	em := NewEquipmentManager()
	assert.NotNil(t, em, "NewEquipmentManager should return non-nil instance")
	assert.NotNil(t, em.slots, "slots map should be initialized")
	assert.Equal(t, 0, len(em.slots), "new manager should have no slots")
}

func TestGetManager_Singleton(t *testing.T) {
	// Reset global manager for test
	globalManager = nil
	once = sync.Once{}

	em1 := GetManager()
	em2 := GetManager()

	assert.NotNil(t, em1, "GetManager should return non-nil instance")
	assert.Same(t, em1, em2, "GetManager should return same instance (singleton)")
}

func TestDefineSlot(t *testing.T) {
	em := NewEquipmentManager()

	// Test valid slot definition
	assert.True(t, em.DefineSlot(1, 2), "DefineSlot should succeed with valid parameters")

	// Test invalid max slots
	assert.False(t, em.DefineSlot(2, 0), "DefineSlot should fail with maxSlots = 0")
	assert.False(t, em.DefineSlot(3, -1), "DefineSlot should fail with negative maxSlots")

	// Test updating existing slot
	assert.True(t, em.EquipItem(1, 100), "Should be able to equip item")
	assert.True(t, em.EquipItem(1, 101), "Should be able to equip second item")
	assert.True(t, em.DefineSlot(1, 1), "Should be able to update existing slot")
	
	items := em.GetEquippedItems(1)
	assert.Equal(t, 1, len(items), "Items should be truncated when reducing MaxSlots")
}

func TestRemoveSlotDefinition(t *testing.T) {
	em := NewEquipmentManager()

	// Test removing non-existent slot
	assert.False(t, em.RemoveSlotDefinition(1), "RemoveSlotDefinition should fail for non-existent slot")

	// Test removing existing slot
	assert.True(t, em.DefineSlot(1, 2), "Should define slot first")
	assert.True(t, em.EquipItem(1, 100), "Should equip item")
	assert.True(t, em.RemoveSlotDefinition(1), "Should successfully remove slot")
}

func TestEquipItem(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot(1, 2)

	// Test valid equip
	assert.True(t, em.EquipItem(1, 100), "Should successfully equip valid item")
	items := em.GetEquippedItems(1)
	assert.Contains(t, items, 100, "Item should be equipped")

	// Test invalid item IDs
	assert.False(t, em.EquipItem(1, 0), "Should fail to equip item ID 0")
	assert.False(t, em.EquipItem(1, -1), "Should fail to equip negative item ID")

	// Test equipping to non-existent slot
	assert.False(t, em.EquipItem(99, 200), "Should fail to equip to non-existent slot")

	// Test duplicate item
	assert.False(t, em.EquipItem(1, 100), "Should fail to equip duplicate item")

	// Test slot capacity
	assert.True(t, em.EquipItem(1, 101), "Should equip second item")
	assert.False(t, em.EquipItem(1, 102), "Should fail when slot is full")
}

func TestUnequipItem(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot(1, 3)
	em.EquipItem(1, 100)
	em.EquipItem(1, 101)
	em.EquipItem(1, 102)

	// Test valid unequip
	assert.True(t, em.UnequipItem(1, 101), "Should successfully unequip existing item")
	items := em.GetEquippedItems(1)
	assert.NotContains(t, items, 101, "Item should be unequipped")
	assert.Contains(t, items, 100, "Other items should remain")
	assert.Contains(t, items, 102, "Other items should remain")

	// Test unequipping non-existent item
	assert.False(t, em.UnequipItem(1, 999), "Should fail to unequip non-existent item")

	// Test unequipping from non-existent slot
	assert.False(t, em.UnequipItem(99, 100), "Should fail to unequip from non-existent slot")
}

func TestGetEquippedItems(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot(1, 3)

	// Test empty slot
	items := em.GetEquippedItems(1)
	assert.Empty(t, items, "Empty slot should return empty slice")

	// Test non-existent slot
	items = em.GetEquippedItems(99)
	assert.Empty(t, items, "Non-existent slot should return empty slice")

	// Test with equipped items
	em.EquipItem(1, 100)
	em.EquipItem(1, 101)
	items = em.GetEquippedItems(1)
	assert.Len(t, items, 2, "Should return correct number of items")
	assert.Contains(t, items, 100, "Should contain first item")
	assert.Contains(t, items, 101, "Should contain second item")

	// Test slice independence (modifying returned slice shouldn't affect internal state)
	items[0] = 999
	originalItems := em.GetEquippedItems(1)
	assert.NotContains(t, originalItems, 999, "Internal state should not be affected by external modifications")
}

func TestIsSlotEmpty(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot(1, 2)

	// Test empty slot
	assert.True(t, em.IsSlotEmpty(1), "Empty slot should return true")

	// Test non-existent slot
	assert.True(t, em.IsSlotEmpty(99), "Non-existent slot should return true")

	// Test slot with items
	em.EquipItem(1, 100)
	assert.False(t, em.IsSlotEmpty(1), "Slot with items should return false")
}

func TestIsSlotFull(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot(1, 2)

	// Test empty slot
	assert.False(t, em.IsSlotFull(1), "Empty slot should not be full")

	// Test partially filled slot
	em.EquipItem(1, 100)
	assert.False(t, em.IsSlotFull(1), "Partially filled slot should not be full")

	// Test full slot
	em.EquipItem(1, 101)
	assert.True(t, em.IsSlotFull(1), "Full slot should return true")

	// Test non-existent slot
	assert.False(t, em.IsSlotFull(99), "Non-existent slot should return false")
}

func TestIsItemEquipped(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot(1, 2)
	em.EquipItem(1, 100)

	// Test equipped item
	assert.True(t, em.IsItemEquipped(1, 100), "Should return true for equipped item")

	// Test non-equipped item
	assert.False(t, em.IsItemEquipped(1, 101), "Should return false for non-equipped item")

	// Test non-existent slot
	assert.False(t, em.IsItemEquipped(99, 100), "Should return false for non-existent slot")
}

func TestGetAllEquippedItems(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot(1, 2)
	em.DefineSlot(2, 1)
	em.DefineSlot(3, 3)

	// Test empty manager
	items := em.GetAllEquippedItems()
	assert.Empty(t, items, "Empty manager should return empty slice")

	// Test with equipped items
	em.EquipItem(1, 100)
	em.EquipItem(1, 101)
	em.EquipItem(2, 200)
	em.EquipItem(3, 300)
	em.EquipItem(3, 301)

	items = em.GetAllEquippedItems()
	expected := []int{100, 101, 200, 300, 301}
	
	// Sort both slices since order might vary
	sort.Ints(items)
	sort.Ints(expected)
	
	assert.Equal(t, expected, items, "Should return all equipped items from all slots")
}

func TestGetAllSlotTypes(t *testing.T) {
	em := NewEquipmentManager()

	// Test empty manager
	slotTypes := em.GetAllSlotTypes()
	assert.Empty(t, slotTypes, "Empty manager should return empty slice")

	// Test with defined slots
	em.DefineSlot(1, 2)
	em.DefineSlot(5, 1)
	em.DefineSlot(10, 3)

	slotTypes = em.GetAllSlotTypes()
	expected := []int{1, 5, 10}
	
	// Sort both slices since order might vary
	sort.Ints(slotTypes)
	sort.Ints(expected)
	
	assert.Equal(t, expected, slotTypes, "Should return all defined slot types")
}

func TestClear(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot(1, 2)
	em.DefineSlot(2, 1)
	em.EquipItem(1, 100)
	em.EquipItem(1, 101)
	em.EquipItem(2, 200)

	// Test clear
	em.Clear()

	// Items should be cleared
	assert.True(t, em.IsSlotEmpty(1), "Slot should be empty after Clear")
	assert.True(t, em.IsSlotEmpty(2), "Slot should be empty after Clear")
	
	items := em.GetAllEquippedItems()
	assert.Empty(t, items, "All items should be cleared")
}

func TestClearSlot(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot(1, 2)
	em.DefineSlot(2, 1)
	em.EquipItem(1, 100)
	em.EquipItem(1, 101)
	em.EquipItem(2, 200)

	// Test clearing specific slot
	assert.True(t, em.ClearSlot(1), "Should successfully clear existing slot")
	assert.True(t, em.IsSlotEmpty(1), "Cleared slot should be empty")
	assert.False(t, em.IsSlotEmpty(2), "Other slots should not be affected")

	// Test clearing non-existent slot
	assert.False(t, em.ClearSlot(99), "Should fail to clear non-existent slot")
}

func TestReset(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot(1, 2)
	em.DefineSlot(2, 1)
	em.EquipItem(1, 100)
	em.EquipItem(2, 200)

	// Test reset
	em.Reset()

	slotTypes := em.GetAllSlotTypes()
	assert.Empty(t, slotTypes, "All slot types should be removed")
	
	items := em.GetAllEquippedItems()
	assert.Empty(t, items, "All items should be removed")
}

func TestGetSlotAvailability(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot(1, 3)

	// Test empty slot
	availability := em.GetSlotAvailability(1)
	assert.Equal(t, 3, availability, "Empty slot should have full availability")

	// Test partially filled slot
	em.EquipItem(1, 100)
	availability = em.GetSlotAvailability(1)
	assert.Equal(t, 2, availability, "Should return correct availability")

	// Test full slot
	em.EquipItem(1, 101)
	em.EquipItem(1, 102)
	availability = em.GetSlotAvailability(1)
	assert.Equal(t, 0, availability, "Full slot should have zero availability")

	// Test non-existent slot
	availability = em.GetSlotAvailability(99)
	assert.Equal(t, 0, availability, "Non-existent slot should return zero availability")
}

func TestConcurrency(t *testing.T) {
	em := NewEquipmentManager()
	em.DefineSlot(1, 100)

	// Test concurrent operations
	var wg sync.WaitGroup
	numGoroutines := 10
	itemsPerGoroutine := 10

	// Concurrent equip operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				itemID := start*itemsPerGoroutine + j + 1
				em.EquipItem(1, itemID)
			}
		}(i)
	}

	wg.Wait()

	// Verify no data races occurred and items were equipped
	items := em.GetEquippedItems(1)
	assert.True(t, len(items) > 0, "Should have equipped some items")
	assert.True(t, len(items) <= 100, "Should not exceed slot capacity")

	// Test concurrent read operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			em.GetAllEquippedItems()
			em.IsSlotEmpty(1)
			em.IsSlotFull(1)
			em.GetSlotAvailability(1)
		}()
	}

	wg.Wait()
	// If we reach here without panic, concurrent reads work correctly
}

func TestEdgeCases(t *testing.T) {
	em := NewEquipmentManager()

	// Test operations on undefined slots
	assert.False(t, em.EquipItem(1, 100), "Should fail to equip to undefined slot")
	assert.False(t, em.UnequipItem(1, 100), "Should fail to unequip from undefined slot")

	// Test with zero max slots after defining
	em.DefineSlot(1, 2)
	em.EquipItem(1, 100)
	em.DefineSlot(1, 0) // This should fail
	items := em.GetEquippedItems(1)
	assert.Len(t, items, 1, "Items should remain when DefineSlot fails")

	// Test large item IDs
	em.DefineSlot(2, 1)
	largeItemID := 999999999
	assert.True(t, em.EquipItem(2, largeItemID), "Should handle large item IDs")
	assert.True(t, em.IsItemEquipped(2, largeItemID), "Should find large item IDs")
}

func TestEquipmentManagerIterator(t *testing.T) {
	em := &EquipmentManager{
		mu: sync.RWMutex{},
		slots: make(map[int]*SlotConfig),
	}
	em.slots[1] = &SlotConfig{ItemIDS: []int{10, 20}, MaxSlots: 2}
	em.slots[2] = &SlotConfig{ItemIDS: []int{30}, MaxSlots: 2}

	em.ResetIterator()

	expected := []int{10, 20, 30}
	var got []int

	for {
		id := em.NextEquippedItem()
		if id == -1 {
			break
		}
		got = append(got, id)
	}

	assert.Equal(t, expected, got, "iterator should return all equipped items in order")
	assert.Equal(t, -1, em.NextEquippedItem(), "iterator should return -1 after exhaustion")

	// Test reset
	em.ResetIterator()
	assert.Equal(t, 10, em.NextEquippedItem(), "first item after reset should be 10")
}