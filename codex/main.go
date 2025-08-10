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
	"codex/pkg/pubsub"
	"sync"
)

var (
    counter int
    mu      sync.Mutex
)

var draggedSlot = inventory.DraggedSlot{Empty: true}
var inv *inventory.Inventory
var em *equipment.EquipmentManager

// C callback function type for message handling
type MessageHandler func(topic *C.char, content *C.char, messageID C.longlong)

var (
	messageHandlers = make(map[int64]MessageHandler)
	handlersMu      sync.RWMutex
)


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
func EquipmentResetIterator(index C.int) C.bool {
	if em == nil {
		return C.bool(0)
	}
	
	em.ResetIterator()
	return C.bool(1)
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
		return C.bool(0)
	}
	
	em.Reset()
	return C.bool(1)
}

//export EquipmentClearSlot
func EquipmentClearSlot(slotType C.int) C.bool {
	if em == nil {
		return C.bool(0)
	}
	
	return C.bool(em.ClearSlot(slotType))
}


//export EquipmentGetSlotAvailability
func EquipmentGetSlotAvailability(slotType C.int) C.bool {
	if em == nil {
		return C.bool(0)
	}
	
	return C.bool(em.GetSlotAvailability(slotType))
}

//export InitializePubSub
func InitializePubSub() {
	InitPubSub()
}

//export PublishMessage
func PublishMessage(topic *C.char, content *C.char) {
	ps := InitPubSub()
	topicStr := C.GoString(topic)
	contentStr := C.GoString(content)
	
	ps.Publish(topicStr, contentStr)
}

//export SubscribeToTopic
func SubscribeToTopic(topic *C.char, handlerID C.longlong) C.longlong {
	ps := InitPubSub()
	topicStr := C.GoString(topic)
	handlerIDInt := int64(handlerID)
	
	// Create a handler that calls the C callback
	handler := func(msg Message) {
		handlersMu.RLock()
		callback, exists := messageHandlers[handlerIDInt]
		handlersMu.RUnlock()
		
		if exists {
			topicC := C.CString(msg.Topic)
			contentC := C.CString(msg.Content)
			defer C.free(unsafe.Pointer(topicC))
			defer C.free(unsafe.Pointer(contentC))
			
			callback(topicC, contentC, C.longlong(msg.ID))
		} else {
			fmt.Printf("Handler %d received message: [%s] %s\n", handlerIDInt, msg.Topic, msg.Content)
		}
	}
	
	listenerID := ps.Subscribe(topicStr, handler)
	fmt.Printf("Subscribed to topic '%s' with listener ID: %d (handler ID: %d)\n", topicStr, listenerID, handlerIDInt)
	
	return C.longlong(listenerID)
}

//export UnsubscribeFromTopic
func UnsubscribeFromTopic(listenerID C.longlong) C.int {
	ps := InitPubSub()
	success := ps.Unsubscribe(int64(listenerID))
	
	if success {
		fmt.Printf("Unsubscribed listener ID: %d\n", int64(listenerID))
		return 1
	}
	
	fmt.Printf("Failed to unsubscribe listener ID: %d\n", int64(listenerID))
	return 0
}

//export RegisterMessageHandler
func RegisterMessageHandler(handlerID C.longlong, callback unsafe.Pointer) {
	// This would be used for more complex C callback scenarios
	// For now, we'll use the simple approach in SubscribeToTopic
	handlerIDInt := int64(handlerID)
	fmt.Printf("Registered message handler ID: %d\n", handlerIDInt)
}

//export GetListenerCount
func GetListenerCount(topic *C.char) C.int {
	ps := InitPubSub()
	topicStr := C.GoString(topic)
	count := ps.GetListenerCount(topicStr)
	return C.int(count)
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
