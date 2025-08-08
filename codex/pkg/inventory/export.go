package inventory

import "C"

//export InventoryNew
func InventoryNew(slotCount int) {
	NewInventory(slotCount)
}

//export InventoryAddItem
func InventoryAddItem(id int, stackable bool, maxStackSize int, qty int) bool {
	return inventory.AddItem(id, stackable, maxStackSize, qty)
}

//export InventoryRemoveItem
func InventoryRemoveItem(id int, qty int) bool {
	return inventory.RemoveItem(id, qty)
}

//export InventoryCountItem
func InventoryCountItem(id int) int {
	return inventory.CountItem(id)
}

//export InventoryPickUpFromSlot
func InventoryPickUpFromSlot(slotIdx int) bool {
	return inventory.PickUpFromSlot(slotIdx)
}

//export InventoryDropToSlot
func InventoryDropToSlot(targetIdx int) bool {
	return inventory.DropToSlot(targetIdx)
}

//export InventoryTakeOneFromSlot
func InventoryTakeOneFromSlot(slotIdx int) bool {
	return inventory.TakeOneFromSlot(slotIdx)
}