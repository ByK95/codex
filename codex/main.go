package main

import "C"
import (
	"codex/pkg/inventory" // module path + folder
	"sync"
)

var (
    counter int
    mu      sync.Mutex
)

var draggedSlot = inventory.DraggedSlot{Empty: true}
var inv *inventory.Inventory

//export InventoryNew
func InventoryNew(slotCount C.int) {
    inv = inventory.NewInventory(int(slotCount))
    draggedSlot = inventory.DraggedSlot{Item: nil, Empty: true}
}

//export InventoryAddItem
func InventoryAddItem(id C.int, stackable C.bool, maxStackSize C.int, qty C.int) C.bool {
    return C.bool(inv.AddItem(int(id), bool(stackable), int(maxStackSize), int(qty)))
}

//export InventoryRemoveItem
func InventoryRemoveItem(id C.int, qty C.int) C.bool {
    return C.bool(inv.RemoveItem(int(id), int(qty)))
}

//export InventoryCountItem
func InventoryCountItem(id C.int) C.int {
    return C.int(inv.CountItem(int(id)))
}

//export InventoryPickUpFromSlot
func InventoryPickUpFromSlot(slotIdx C.int) C.bool {
    return C.bool(inv.PickUpFromSlot(draggedSlot, int(slotIdx)))
}

//export InventoryDropToSlot
func InventoryDropToSlot(targetIdx C.int) C.bool {
    return C.bool(inv.DropToSlot(draggedSlot, int(targetIdx)))
}

//export InventoryTakeOneFromSlot
func InventoryTakeOneFromSlot(slotIdx C.int) C.bool {
    return C.bool(inv.TakeOneFromSlot(draggedSlot, int(slotIdx)))
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
