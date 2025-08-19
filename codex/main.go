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
	"codex/pkg/metrics"
	"codex/pkg/zone"
	"codex/pkg/store"
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

func Metrics_IncInt(name *C.char) {
	metrics.IncInt(C.GoString(name))
}

//export Metrics_AddInt
func Metrics_AddInt(name *C.char, val C.longlong) {
	metrics.AddInt(C.GoString(name), int64(val))
}

//export Metrics_GetInt
func Metrics_GetInt(name *C.char) C.longlong {
	return C.longlong(metrics.GetInt(C.GoString(name)))
}

// Float metrics
//export Metrics_AddFloat
func Metrics_AddFloat(name *C.char, val C.double) {
	metrics.AddFloat(C.GoString(name), float64(val))
}

//export Metrics_GetFloat
func Metrics_GetFloat(name *C.char) C.double {
	return C.double(metrics.GetFloat(C.GoString(name)))
}

// Bool metrics
//export Metrics_SetBool
func Metrics_SetBool(name *C.char, val C.int) {
	metrics.SetBool(C.GoString(name), val != 0)
}

//export Metrics_GetBool
func Metrics_GetBool(name *C.char) C.int {
	if metrics.GetBool(C.GoString(name)) {
		return 1
	}
	return 0
}

// String metrics
//export Metrics_SetString
func Metrics_SetString(name *C.char, val *C.char) {
	metrics.SetString(C.GoString(name), C.GoString(val))
}

//export Metrics_GetString
func Metrics_GetString(name *C.char) *C.char {
	s := metrics.GetString(C.GoString(name))
	return C.CString(s)
}

//export Metrics_SnapshotJSON
func Metrics_SnapshotJSON() *C.char {
	s := metrics.SnapshotJSON()
	return C.CString(s)
}

//export Metrics_FreeCString
func Metrics_FreeCString(str *C.char) {
	C.free(unsafe.Pointer(str))
}

//export Metrics_ClearAll
func Metrics_ClearAll() {
	metrics.ClearAll()
}

//export Metrics_ClearPrefix
func Metrics_ClearPrefix(prefix *C.char) {
	metrics.ClearPrefix(C.GoString(prefix))
}

//export Metrics_LoadFromJSON
func Metrics_LoadFromJSON(jsonStr *C.char) C.int {
	if err := metrics.LoadFromJSON(C.GoString(jsonStr)); err != nil {
		return 1 // error
	}
	return 0 // success
}

// Loot functions are not added reason being in order to pass the list into unreal and back needs 3 different conversations which doesn't worth it imo also taken a look into flatbuffers which offer nice conversations third will be necessary for blueprints again so given up atm but leaving the logic here

//export IncreaseThreat
func IncreaseThreat(zoneID C.int, amount C.float) C.float {
	return C.float(threat.GetManager().IncreaseThreat(int32(zoneID), float32(amount)))
}

//export TimedThreat
func TimedThreat(currentID C.int, amount C.float) C.float {
	return C.float(threat.GetManager().TimedThreat(int32(currentID), float32(amount)))
}

//export Threat_AdvanceMap
func Threat_AdvanceMap() {
	threat.GetManager().AdvanceMap()
}

//export GetThreat
func GetThreat(zoneID C.int) C.float {
	return C.float(threat.GetManager().GetZoneThreat(int32(zoneID)))
}

//export ResetThreats
func ResetThreats() {
	threat.GetManager().Reset()
}

// Optional: expose a string for debugging
//export ThreatManagerString
func ThreatManagerString() *C.char {
	s := threat.GetManager().String()
	return C.CString(s)
}


// Optional: expose a string for debugging
//export ZoneManagerString
func ZoneManagerString() *C.char {
	s := zone.GetManager().String()
	return C.CString(s)
}

// ---- Int ----

//export Store_SetInt
func Store_SetInt(key *C.char, val C.longlong) {
	GetStore().SetInt(C.GoString(key), int64(val))
}

//export Store_GetInt
func Store_GetInt(key *C.char) C.longlong {
	return C.longlong(GetStore().GetInt(C.GoString(key)))
}

//export Store_AddInt
func Store_AddInt(key *C.char, val C.longlong) {
	GetStore().AddInt(C.GoString(key), int64(val))
}

//export Store_SubInt
func Store_SubInt(key *C.char, val C.longlong) C.bool {
	return C.bool(GetStore().SubInt(C.GoString(key), int64(val)))
}

// ---- Float ----

//export Store_SetFloat
func Store_SetFloat(key *C.char, val C.double) {
	store.GetStore().SetFloat(C.GoString(key), float64(val))
}

//export Store_GetFloat
func Store_GetFloat(key *C.char) C.double {
	return C.double(store.GetStore().GetFloat(C.GoString(key)))
}

//export Store_AddFloat
func Store_AddFloat(key *C.char, val C.double) {
	store.GetStore().AddFloat(C.GoString(key), float64(val))
}

//export Store_SubFloat
func Store_SubFloat(key *C.char, val C.double) C.bool {
	return C.bool(store.GetStore().SubFloat(C.GoString(key), float64(val)))
}

// ---- Bool ----

//export Store_SetBool
func Store_SetBool(key *C.char, val C.bool) {
	store.GetStore().SetBool(C.GoString(key), bool(val))
}

//export Store_GetBool
func Store_GetBool(key *C.char) C.bool {
	return C.bool(store.GetStore().GetBool(C.GoString(key)))
}

// ---- String ----

//export Store_SetString
func Store_SetString(key *C.char, val *C.char) {
	store.GetStore().SetString(C.GoString(key), C.GoString(val))
}

//export Store_GetString
func Store_GetString(key *C.char) *C.char {
	s := store.GetStore().GetString(C.GoString(key))
	return C.CString(s) // caller must free
}

// ---- Persistence ----

//export Store_Save
func Store_Save() C.int {
	if err := store.GetStore().Save(); err != nil {
		return -1
	}
	return 0
}

//export Store_Load
func Store_Load() C.int {
	if err := store.GetStore().Load(); err != nil {
		return -1
	}
	return 0
}

//export ZoneConfig_GetMaxNPC
func ZoneConfig_GetMaxNPC(zoneID *C.char) C.int {
	id := C.GoString(zoneID)
	return C.int(GetManager("zones.json").GetMaxNPC(id))
}

//export ZoneConfig_GetSpawnChance
func ZoneConfig_GetSpawnChance(zoneID *C.char) C.float {
	id := C.GoString(zoneID)
	return C.float(GetManager("zones.json").GetSpawnChance(id))
}

//export ZoneConfig_GetRandomNPCType
func ZoneConfig_GetRandomNPCType(zoneID *C.char) *C.char {
	id := C.GoString(zoneID)
	npc := GetManager("zones.json").GetRandomNPCType(id)
	return C.CString(npc)
}

//export ZoneConfig_Reload
func ZoneConfig_Reload() C.int {
	err := GetManager("zones.json").Load()
	if err != nil {
		return 0
	}
	return 1
}

func main() {}
