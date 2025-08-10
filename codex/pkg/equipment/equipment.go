// Package equipment provides a flexible equipment slot management system
package equipment

import (
	"sort"
	"sync"
)

// SlotConfig defines configuration for an equipment slot
type SlotConfig struct {
	ItemIDS  []int // List of currently equipped item IDs (empty slice = no items)
	MaxSlots int   // Maximum number of items that can be equipped in this slot type
}

// EquipmentManager manages equipment slots with runtime-configurable slot types
type EquipmentManager struct {
	mu    sync.RWMutex
	slots map[int]*SlotConfig // slotType -> SlotConfig
	iterIndex    int
	iterItems    []int
}

// Global equipment manager instance
var globalManager *EquipmentManager
var once sync.Once

// GetManager returns the global equipment manager instance (singleton)
func GetManager() *EquipmentManager {
	once.Do(func() {
		globalManager = NewEquipmentManager()
	})
	return globalManager
}

// NewEquipmentManager creates a new equipment manager
func NewEquipmentManager() *EquipmentManager {
	return &EquipmentManager{
		slots: make(map[int]*SlotConfig),
		iterIndex: 0,
	}
}

// DefineSlot defines a new slot type with its maximum capacity
// If the slot type already exists, it updates the MaxSlots but keeps the current ItemIDS
// Returns true on success, false on failure
func (em *EquipmentManager) DefineSlot(slotType int, maxSlots int) bool {
	if maxSlots < 1 {
		return false
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	if existing, exists := em.slots[slotType]; exists {
		// Update max slots but keep current items
		existing.MaxSlots = maxSlots
		// If the new max is smaller than current items, truncate the list
		if len(existing.ItemIDS) > maxSlots {
			existing.ItemIDS = existing.ItemIDS[:maxSlots]
		}
	} else {
		// Create new slot
		em.slots[slotType] = &SlotConfig{
			ItemIDS:  make([]int, 0), // Empty by default
			MaxSlots: maxSlots,
		}
	}

	return true
}

// RemoveSlotDefinition removes a slot type definition entirely
// Returns true on success, false if slot doesn't exist
func (em *EquipmentManager) RemoveSlotDefinition(slotType int) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.slots[slotType]; !exists {
		return false
	}

	delete(em.slots, slotType)
	return true
}

// EquipItem equips an item to the specified slot type
// Returns true on success, false on failure
func (em *EquipmentManager) EquipItem(slotType int, itemID int) bool {
	if itemID <= 0 { // Changed to <= 0 to prevent 0 as a valid item ID
		return false
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	slot, exists := em.slots[slotType]
	if !exists {
		return false
	}

	// Check if we've reached the maximum slots
	if len(slot.ItemIDS) >= slot.MaxSlots {
		return false
	}

	// Check if item is already equipped in this slot
	for _, id := range slot.ItemIDS {
		if id == itemID {
			return false // Item already equipped
		}
	}

	slot.ItemIDS = append(slot.ItemIDS, itemID)
	return true
}

// UnequipItem removes the item from the specified slot type
// Returns true on success, false if slot doesn't exist or item not found
func (em *EquipmentManager) UnequipItem(slotType int, itemID int) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	slot, exists := em.slots[slotType]
	if !exists {
		return false
	}

	// Find and remove the item
	for i, id := range slot.ItemIDS {
		if id == itemID {
			slot.ItemIDS = append(slot.ItemIDS[:i], slot.ItemIDS[i+1:]...)
			return true
		}
	}

	return false // Item not found
}

// GetEquippedItems returns all item IDs equipped in the specified slot type
// Returns empty slice if slot is empty or doesn't exist
func (em *EquipmentManager) GetEquippedItems(slotType int) []int {
	em.mu.RLock()
	defer em.mu.RUnlock()

	slot, exists := em.slots[slotType]
	if !exists {
		return []int{}
	}

	// Return a copy to prevent external modification
	result := make([]int, len(slot.ItemIDS))
	copy(result, slot.ItemIDS)
	return result
}

// IsSlotEmpty returns true if the specified slot type is empty or doesn't exist
func (em *EquipmentManager) IsSlotEmpty(slotType int) bool {
	items := em.GetEquippedItems(slotType)
	return len(items) == 0
}

// IsSlotFull returns true if the specified slot type has reached its maximum capacity
func (em *EquipmentManager) IsSlotFull(slotType int) bool {
	em.mu.RLock()
	defer em.mu.RUnlock()

	slot, exists := em.slots[slotType]
	if !exists {
		return false
	}

	return len(slot.ItemIDS) >= slot.MaxSlots
}

// IsItemEquipped returns true if the specified item is equipped in the given slot type
func (em *EquipmentManager) IsItemEquipped(slotType int, itemID int) bool {
	items := em.GetEquippedItems(slotType)
	for _, id := range items {
		if id == itemID {
			return true
		}
	}
	return false
}

// GetAllEquippedItems returns an array of all equipped item IDs across all slots
func (em *EquipmentManager) GetAllEquippedItems() []int {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var result []int
	for _, slot := range em.slots {
		result = append(result, slot.ItemIDS...)
	}

	return result
}

// ResetIterator resets the iterator state
func (em *EquipmentManager) ResetIterator() {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.iterIndex = 0
	em.iterItems = nil
	for _, slot := range em.slots {
		em.iterItems = append(em.iterItems, slot.ItemIDS...)
	}
	sort.Ints(em.iterItems)
}

// NextEquippedItem returns the next equipped item ID or -1 if finished
func (em *EquipmentManager) NextEquippedItem() int {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if em.iterIndex >= len(em.iterItems) {
		return -1
	}
	val := em.iterItems[em.iterIndex]
	em.iterIndex++
	return val
}

// GetAllSlotTypes returns all defined slot types
func (em *EquipmentManager) GetAllSlotTypes() []int {
	em.mu.RLock()
	defer em.mu.RUnlock()

	slotTypes := make([]int, 0, len(em.slots))
	for slotType := range em.slots {
		slotTypes = append(slotTypes, slotType)
	}

	return slotTypes
}

// Clear removes all equipped items but keeps slot definitions
func (em *EquipmentManager) Clear() {
	em.mu.Lock()
	defer em.mu.Unlock()

	for _, slot := range em.slots {
		slot.ItemIDS = slot.ItemIDS[:0] // Clear slice but keep capacity
	}
}

// ClearSlot removes all equipped items from a specific slot but keeps the slot definition
func (em *EquipmentManager) ClearSlot(slotType int) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	slot, exists := em.slots[slotType]
	if !exists {
		return false
	}

	slot.ItemIDS = slot.ItemIDS[:0] // Clear slice but keep capacity
	return true
}

// Reset removes all slot definitions and equipped items
func (em *EquipmentManager) Reset() {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.slots = make(map[int]*SlotConfig)
}

// GetSlotCapacity returns the current usage and maximum capacity of a slot
// Returns (currentCount, maxSlots, exists)
func (em *EquipmentManager) GetSlotAvailability(slotType int) int  {
	em.mu.RLock()
	defer em.mu.RUnlock()

	slot, exists := em.slots[slotType]
	if !exists {
		return 0
	}

	return slot.MaxSlots - len(slot.ItemIDS)
}