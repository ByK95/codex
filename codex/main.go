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
	"codex/pkg/threat"
	"codex/pkg/store"
	"codex/pkg/zoneconfig"
	voronoi "codex/pkg/grid_voronoi"
	"codex/pkg/crafting"
	"codex/pkg/helpers"
	"sync"
	"codex/pkg/storage"
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

//export EquipmentDefineSlot
func EquipmentDefineSlot(slotType *C.char, maxSlots C.int) C.int {
	if equipment.GetManager().DefineSlot(C.GoString(slotType), int(maxSlots)) {
		return 1
	}
	return 0
}

//export EquipmentRemoveSlotDefinition
func EquipmentRemoveSlotDefinition(slotType *C.char) C.int {
	if equipment.GetManager().RemoveSlotDefinition(C.GoString(slotType)) {
		return 1
	}
	return 0
}

//export EquipmentEquipItem
func EquipmentEquipItem(slotType *C.char, itemID *C.char) C.int {
	if equipment.GetManager().EquipItem(C.GoString(slotType), C.GoString(itemID)) {
		return 1
	}
	return 0
}

//export EquipmentUnequipItem
func EquipmentUnequipItem(slotType *C.char, itemID *C.char) C.int {
	if equipment.GetManager().UnequipItem(C.GoString(slotType), C.GoString(itemID)) {
		return 1
	}
	return 0
}

//export EquipmentIsSlotFull
func EquipmentIsSlotFull(slotType *C.char) C.bool {
	return C.bool(equipment.GetManager().IsSlotFull(C.GoString(slotType)))
}

//export EquipmentIsItemEquipped
func EquipmentIsItemEquipped(slotType *C.char, itemID *C.char) C.bool {
	em := equipment.GetManager()
	return C.bool(em.IsItemEquipped(C.GoString(slotType), C.GoString(itemID)))
}

//export InitGetAllEquipmentItemsIter
func InitGetAllEquipmentItemsIter() C.bool {
	equipment.InitGetAllEquippedItemsIter()
	return true
}

//export InitGetAllEquipmentSlotsIter
func InitGetAllEquipmentSlotsIter() C.bool {
	equipment.InitGetAllSlotsIter()
	return true
}

//export EquipmentNext
func EquipmentNext() *C.char {
	item := equipment.Next()
	return C.CString(item)
}

//export EquipmentClearSlot
func EquipmentClearSlot(slotType *C.char) C.bool {
	em := equipment.GetManager()
	return C.bool(em.ClearSlot(C.GoString(slotType)))
}

//export EquipmentGetSlotAvailability
func EquipmentGetSlotAvailability(slotType *C.char) C.int {
	em := equipment.GetManager()
	return C.int(em.GetSlotAvailability(C.GoString(slotType)))
}

//export EquipmentClear
func EquipmentClear() {
	equipment.Clear()
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
	s, _ := metrics.SnapshotJSON()
	str := string(s.([]byte))
	return C.CString(str)
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

//export Threat_GetMapFactor
func Threat_GetMapFactor() C.float {
	return C.float(threat.GetManager().GetMapFactor())
}

//export GetThreat
func GetThreat(zoneID C.int) C.float {
	return C.float(threat.GetManager().GetZoneThreat(int32(zoneID)))
}

//export ResetThreats
func ResetThreats() {
	threat.GetManager().Reset()
}

//export Store_RandomSelect
func Store_RandomSelect(prefix *C.char) *C.char {
	res := store.GetStore().RandomSelect(C.GoString(prefix))
	return C.CString(res)
}

//export Store_SetInt
func Store_SetInt(key *C.char, val C.longlong) {
	store.GetStore().SetInt(C.GoString(key), int64(val))
}

//export Store_GetInt
func Store_GetInt(key *C.char) C.longlong {
	return C.longlong(store.GetStore().GetInt(C.GoString(key)))
}

//export Store_AddInt
func Store_AddInt(key *C.char, val C.longlong) {
	store.GetStore().AddInt(C.GoString(key), int64(val))
}

//export Store_SubInt
func Store_SubInt(key *C.char, val C.longlong) C.bool {
	return C.bool(store.GetStore().SubInt(C.GoString(key), int64(val)))
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

//export Store_ReleaseBool
func Store_ReleaseBool(key *C.char) C.bool {
	return C.bool(store.GetStore().ReleaseBool(C.GoString(key)))
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

//export Storage_Save
func Storage_Save() *C.char {
	if err := storage.SM().SaveAll(); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export Storage_Load
func Storage_Load() *C.char {
	if err := storage.SM().LoadAll(); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export Storage_ReloadAll
func Storage_ReloadAll() *C.char {
	if err := storage.SM().ReloadAll(); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export SetStorageManagerPath
func SetStorageManagerPath(path *C.char) *C.char {
	p := C.GoString(path)
	if err := storage.SetStorageManagerPath(p); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export ZoneConfig_GetMaxNPC
func ZoneConfig_GetMaxNPC(zoneID *C.char) C.int {
	id := C.GoString(zoneID)
	return C.int(zoneconfig.GetManager().GetMaxNPC(id))
}

//export ZoneConfig_GetSpawnChance
func ZoneConfig_GetSpawnChance(zoneID *C.char) C.float {
	id := C.GoString(zoneID)
	return C.float(zoneconfig.GetManager().GetSpawnChance(id))
}

//export ZoneConfig_GetRandomNPCType
func ZoneConfig_GetRandomNPCType(zoneID *C.char) *C.char {
	id := C.GoString(zoneID)
	npc := zoneconfig.GetManager().GetRandomNPCType(id)
	return C.CString(npc)
}

//export Voronoi_Init
func Voronoi_Init(width C.int, height C.int, numZones C.int, seed C.longlong) {
	voronoi.Init(int(width), int(height), int(numZones), int64(seed))
}

//export Voronoi_ZoneAt
func Voronoi_ZoneAt(x C.int, y C.int) C.int {
	return C.int(voronoi.ZoneAt(int(x), int(y)))
}

//export Crafting_Reset
func Crafting_Reset(name *C.char) {
	n := C.GoString(name)
	
	crafting.Reset(n)
}

//export Crafting_ResetAll
func Crafting_ResetAll() {
	crafting.ResetAll()
}

//export Crafting_FindFirstByRequirement
func Crafting_FindFirstByRequirement(managerName *C.char, reqID *C.char) *C.char {
	name := C.GoString(managerName)
	req := C.GoString(reqID)

	m, ok := crafting.Get(name)
	if !ok {
		return C.CString("")
	}

	items := m.FindByRequirement(req)
	if len(items) == 0 {
		return C.CString("")
	}

	return C.CString(items[0].ID)
}

//export Crafting_GetFirstRequirement
func Crafting_GetFirstRequirement(managerName *C.char, craftID *C.char) *C.char {
	name := C.GoString(managerName)
	id := C.GoString(craftID)

	m, ok := crafting.Get(name)
	if !ok {
		return C.CString("")
	}

	c, exists := m.GetCraftable(id)
	if !exists || len(c.Requirements) == 0 {
		return C.CString("")
	}

	return C.CString(c.Requirements[0].ID)
}

//export Helpers_GetUpgradeSelections
func Helpers_GetUpgradeSelections(count C.int) C.bool {
	helpers.GetUpgradeSelections(int(count))
	return true
}

//export Helpers_Next
func Helpers_Next() *C.char {
	item := helpers.GetNextSelections()
	return C.CString(item)
}

//export Store_InitGetFullKeysIter
func Store_InitGetFullKeysIter(prefix *C.char) C.bool {
	p := C.GoString(prefix)
	store.InitGetFullKeysIter(p)
	return true
}

//export Store_Clear
func Store_Clear(prefix *C.char) C.bool {
	p := C.GoString(prefix)
	store.GetStore().Clear(p)
	return true
}

//export Store_Next
func Store_Next() *C.char {
	item := store.Next()
	return C.CString(item)
}


func main() {}
