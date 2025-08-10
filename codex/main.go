package main

/*
#include <stdlib.h>
#include <stdbool.h>
*/
import "C"

import "unsafe"
import (
	"codex/pkg/inventory"
	"codex/pkg/equipment"
	"sync"
)

var (
    counter int
    mu      sync.Mutex
)

var draggedSlot = inventory.DraggedSlot{Empty: true}
var inv *inventory.Inventory
var em *equipment.EquipmentManager

//export InventoryNew
func InventoryNew(slotCount C.int) {
    inv = inventory.NewInventory(int(slotCount))
    draggedSlot = inventory.DraggedSlot{Item: nil, Empty: true}
}

//export InventoryAddItem
func InventoryAddItem(id C.int, stackable C.int, maxStackSize C.int, qty C.int) C.int {
    if inv.AddItem(int(id), stackable != 0, int(maxStackSize), int(qty)) {
        return 1
    }
    return 0
}

//export InventoryRemoveItem
func InventoryRemoveItem(id C.int, qty C.int) C.int {
    if inv.RemoveItem(int(id), int(qty)) {
        return 1
    }
    return 0
}

//export InventoryCountItem
func InventoryCountItem(id C.int) C.int {
    return C.int(inv.CountItem(int(id)))
}

//export InventoryPickUpFromSlot
func InventoryPickUpFromSlot(slotIdx C.int) C.int {
    if inv.PickUpFromSlot(&draggedSlot, int(slotIdx)) {
        return 1
    }
    return 0
}

//export InventoryDropToSlot
func InventoryDropToSlot(targetIdx C.int) C.int {
    if inv.DropToSlot(&draggedSlot, int(targetIdx)) {
        return 1
    }
    return 0
}

//export InventoryTakeOneFromSlot
func InventoryTakeOneFromSlot(slotIdx C.int) C.int {
    if inv.TakeOneFromSlot(&draggedSlot, int(slotIdx)) {
        return 1
    }
    return 0
}

//export InventoryGetSlotItemID
func InventoryGetSlotItemID(slotIdx C.int) C.int {
	if inv == nil {
		return -1 // Invalid state
	}
	
	idx := int(slotIdx)
	if idx < 0 || idx >= len(inv.Slots) {
		return -1 // Invalid slot index
	}
	
	slot := inv.Slots[idx]
	if slot == nil {
		return -1 // Empty slot
	}
	
	return C.int(slot.ID)
}

//export InventoryGetSlotQuantity
func InventoryGetSlotQuantity(slotIdx C.int) C.int {
	if inv == nil {
		return 0
	}
	
	idx := int(slotIdx)
	if idx < 0 || idx >= len(inv.Slots) {
		return 0
	}
	
	slot := inv.Slots[idx]
	if slot == nil {
		return 0
	}
	
	return C.int(slot.Quantity)
}

//export InventoryGetSlotStackable
func InventoryGetSlotStackable(slotIdx C.int) C.bool {
	if inv == nil {
		return false
	}
	
	idx := int(slotIdx)
	if idx < 0 || idx >= len(inv.Slots) {
		return false
	}
	
	slot := inv.Slots[idx]
	if slot == nil {
		return false
	}
	
	return C.bool(slot.Stackable)
}

//export InventoryGetSlotMaxStackSize
func InventoryGetSlotMaxStackSize(slotIdx C.int) C.int {
	if inv == nil {
		return 0
	}
	
	idx := int(slotIdx)
	if idx < 0 || idx >= len(inv.Slots) {
		return 0
	}
	
	slot := inv.Slots[idx]
	if slot == nil {
		return 0
	}
	
	return C.int(slot.MaxStackSize)
}

//export InventoryIsSlotEmpty
func InventoryIsSlotEmpty(slotIdx C.int) C.bool {
	if inv == nil {
		return true
	}
	
	idx := int(slotIdx)
	if idx < 0 || idx >= len(inv.Slots) {
		return true // Invalid slots are considered empty
	}
	
	slot := inv.Slots[idx]
	return C.bool(slot == nil || slot.Quantity == 0)
}

//export EquipmentNew
func EquipmentNew() {
	em = equipment.NewEquipmentManager()
}

//export EquipmentDefineSlot
func EquipmentDefineSlot(slotType C.int, maxSlots C.int) C.int {
	if em == nil {
		return 0
	}
	if em.DefineSlot(int(slotType), int(maxSlots)) {
		return 1
	}
	return 0
}

//export EquipmentRemoveSlotDefinition
func EquipmentRemoveSlotDefinition(slotType C.int) C.int {
	if em == nil {
		return 0
	}
	if em.RemoveSlotDefinition(int(slotType)) {
		return 1
	}
	return 0
}

//export EquipmentEquipItem
func EquipmentEquipItem(slotType C.int, itemID C.int) C.int {
	if em == nil {
		return 0
	}
	if em.EquipItem(int(slotType), int(itemID)) {
		return 1
	}
	return 0
}

//export EquipmentUnequipItem
func EquipmentUnequipItem(slotType C.int, itemID C.int) C.int {
	if em == nil {
		return 0
	}
	if em.UnequipItem(int(slotType), int(itemID)) {
		return 1
	}
	return 0
}

//export EquipmentIsSlotFull
func EquipmentIsSlotFull(slotType C.int) C.bool {
	if em == nil {
		return false
	}
	return C.bool(em.IsSlotFull(int(slotType)))
}

//export EquipmentIsItemEquipped
func EquipmentIsItemEquipped(slotType C.int, itemID C.int) C.bool {
	if em == nil {
		return false
	}
	return C.bool(em.IsItemEquipped(int(slotType), int(itemID)))
}

//export EquipmentResetIterator
func EquipmentResetIterator() C.bool {
	if em == nil {
		return false
	}
	
	em.ResetIterator()
	return true
}

//export EquipmentNextEquippedItem
func EquipmentNextEquippedItem() C.int {
	if em == nil {
		return 0
	}
	
	return C.int(em.NextEquippedItem())
}

//export EquipmentReset
func EquipmentReset() C.bool {
	if em == nil {
		return false
	}
	
	em.Reset()
	return true
}

//export EquipmentClearSlot
func EquipmentClearSlot(slotType C.int) C.bool {
	if em == nil {
		return false
	}
	
	return C.bool(em.ClearSlot(int(slotType)))
}


//export EquipmentGetSlotAvailability
func EquipmentGetSlotAvailability(slotType C.int) C.int {
	if em == nil {
		return 0
	}
	
	return C.int(em.GetSlotAvailability(int(slotType)))
}

//export Increment
func Increment() {
    mu.Lock()
    counter++
    mu.Unlock()
}

//export GetCount
func GetCount() int {
    mu.Lock()
    defer mu.Unlock()
    return counter
}

func main() {}
