// Package equipment provides a flexible equipment slot management system
package equipment

import (
	"codex/pkg/iterator"
	"sort"
	"sync"
)

// SlotConfig defines configuration for an equipment slot
type SlotConfig struct {
	ItemIDS  []string // List of currently equipped item IDs (empty slice = no items)
	MaxSlots int      // Maximum number of items that can be equipped in this slot type
}

// EquipmentManager manages equipment slots with runtime-configurable slot types
type EquipmentManager struct {
	mu        sync.RWMutex
	slots     map[string]*SlotConfig // slotType -> SlotConfig
	iterIndex int
	iterItems []string
}

// Global equipment manager instance
var globalManager *EquipmentManager
var once sync.Once
var equipmentIter *iterator.Iterator[string]

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
		slots:     make(map[string]*SlotConfig),
		iterIndex: 0,
	}
}

// DefineSlot defines a new slot type with its maximum capacity
// If the slot type already exists, it updates the MaxSlots but keeps the current ItemIDS
// Returns true on success, false on failure
func (em *EquipmentManager) DefineSlot(slotType string, maxSlots int) bool {
	if maxSlots < 1 {
		return false
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	if existing, exists := em.slots[slotType]; exists {
		// Update max slots but keep current items
		existing.MaxSlots = maxSlots
		if len(existing.ItemIDS) > maxSlots {
			existing.ItemIDS = existing.ItemIDS[:maxSlots]
		}
	} else {
		em.slots[slotType] = &SlotConfig{
			ItemIDS:  make([]string, 0),
			MaxSlots: maxSlots,
		}
	}

	return true
}

// RemoveSlotDefinition removes a slot type definition entirely
func (em *EquipmentManager) RemoveSlotDefinition(slotType string) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.slots[slotType]; !exists {
		return false
	}

	delete(em.slots, slotType)
	return true
}

// EquipItem equips an item to the specified slot type
func (em *EquipmentManager) EquipItem(slotType string, itemID string) bool {
	if itemID == "" {
		return false
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	slot, exists := em.slots[slotType]
	if !exists {
		return false
	}

	if len(slot.ItemIDS) >= slot.MaxSlots {
		return false
	}

	for _, id := range slot.ItemIDS {
		if id == itemID {
			return false // already equipped
		}
	}

	slot.ItemIDS = append(slot.ItemIDS, itemID)
	return true
}

// UnequipItem removes the item from the specified slot type
func (em *EquipmentManager) UnequipItem(slotType string, itemID string) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	slot, exists := em.slots[slotType]
	if !exists {
		return false
	}

	for i, id := range slot.ItemIDS {
		if id == itemID {
			slot.ItemIDS = append(slot.ItemIDS[:i], slot.ItemIDS[i+1:]...)
			return true
		}
	}

	return false
}

// GetEquippedItems returns all item IDs for a slot
func (em *EquipmentManager) GetEquippedItems(slotType string) []string {
	em.mu.RLock()
	defer em.mu.RUnlock()

	slot, exists := em.slots[slotType]
	if !exists {
		return []string{}
	}

	result := make([]string, len(slot.ItemIDS))
	copy(result, slot.ItemIDS)
	return result
}

// IsSlotEmpty returns true if no items are equipped
func (em *EquipmentManager) IsSlotEmpty(slotType string) bool {
	return len(em.GetEquippedItems(slotType)) == 0
}

// IsSlotFull returns true if slot reached max capacity
func (em *EquipmentManager) IsSlotFull(slotType string) bool {
	em.mu.RLock()
	defer em.mu.RUnlock()

	slot, exists := em.slots[slotType]
	if !exists {
		return false
	}

	return len(slot.ItemIDS) >= slot.MaxSlots
}

// IsItemEquipped checks if item is in slot
func (em *EquipmentManager) IsItemEquipped(slotType string, itemID string) bool {
	for _, id := range em.GetEquippedItems(slotType) {
		if id == itemID {
			return true
		}
	}
	return false
}

// GetAllEquippedItems returns all items across slots
func (em *EquipmentManager) GetAllEquippedItems() []string {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var result []string
	for _, slot := range em.slots {
		result = append(result, slot.ItemIDS...)
	}

	return result
}

// ResetIterator resets item iterator
func (em *EquipmentManager) ResetIterator() {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.iterIndex = 0
	em.iterItems = nil
	for _, slot := range em.slots {
		em.iterItems = append(em.iterItems, slot.ItemIDS...)
	}
	sort.Strings(em.iterItems)
}

// NextEquippedItem returns next item, or "" if finished
func (em *EquipmentManager) NextEquippedItem() string {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if em.iterIndex >= len(em.iterItems) {
		return ""
	}
	val := em.iterItems[em.iterIndex]
	em.iterIndex++
	return val
}

// GetAllSlotTypes returns defined slot types
func (em *EquipmentManager) GetAllSlotTypes() []string {
	em.mu.RLock()
	defer em.mu.RUnlock()

	slotTypes := make([]string, 0, len(em.slots))
	for slotType := range em.slots {
		slotTypes = append(slotTypes, slotType)
	}

	return slotTypes
}

// Clear removes all equipped items
func (em *EquipmentManager) Clear() {
	em.mu.Lock()
	defer em.mu.Unlock()

	for _, slot := range em.slots {
		slot.ItemIDS = slot.ItemIDS[:0]
	}
}

// ClearSlot removes all items from a slot
func (em *EquipmentManager) ClearSlot(slotType string) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	slot, exists := em.slots[slotType]
	if !exists {
		return false
	}

	slot.ItemIDS = slot.ItemIDS[:0]
	return true
}

// Reset removes all slots
func (em *EquipmentManager) Reset() {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.slots = make(map[string]*SlotConfig)
}

// GetSlotAvailability returns remaining capacity
func (em *EquipmentManager) GetSlotAvailability(slotType string) int {
	em.mu.RLock()
	defer em.mu.RUnlock()

	slot, exists := em.slots[slotType]
	if !exists {
		return 0
	}

	return slot.MaxSlots - len(slot.ItemIDS)
}

// HasAnyEmptySlot returns true if at least one defined slot has available capacity
func (em *EquipmentManager) HasAnyEmptySlot() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()

	for _, slot := range em.slots {
		if len(slot.ItemIDS) < slot.MaxSlots {
			return true
		}
	}
	return false
}

func InitGetAllEquippedItemsIter(){
	allItems := GetManager().GetAllEquippedItems()
	equipmentIter = iterator.NewIterator(allItems)
}

func InitGetAllSlotsIter(){
	allSlotTypes := GetManager().GetAllSlotTypes()
	equipmentIter = iterator.NewIterator(allSlotTypes)
}

func Next() string{
	if equipmentIter == nil {
		return ""
	}
	val, _ := equipmentIter.Next()
	return val
}

func Clear(){
	globalManager = NewEquipmentManager()
	equipmentIter = nil
}