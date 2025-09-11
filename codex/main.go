package main

/*
#include <stdlib.h>
#include <stdbool.h>

typedef struct {
    int x;
    int y;
} Coord2d;

static inline Coord2d MakeCoord(int x, int y) {
    Coord2d c;
    c.x = x;
    c.y = y;
    return c;
}
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
func EquipmentDefineSlot(slotType *C.char, maxSlots C.int) C.int {
	if em == nil {
		return 0
	}
	if em.DefineSlot(C.GoString(slotType), int(maxSlots)) {
		return 1
	}
	return 0
}

//export EquipmentRemoveSlotDefinition
func EquipmentRemoveSlotDefinition(slotType *C.char) C.int {
	if em == nil {
		return 0
	}
	if em.RemoveSlotDefinition(C.GoString(slotType)) {
		return 1
	}
	return 0
}

//export EquipmentEquipItem
func EquipmentEquipItem(slotType *C.char, itemID *C.char) C.int {
	if em == nil {
		return 0
	}
	if em.EquipItem(C.GoString(slotType), C.GoString(itemID)) {
		return 1
	}
	return 0
}

//export EquipmentUnequipItem
func EquipmentUnequipItem(slotType *C.char, itemID *C.char) C.int {
	if em == nil {
		return 0
	}
	if em.UnequipItem(C.GoString(slotType), C.GoString(itemID)) {
		return 1
	}
	return 0
}

//export EquipmentIsSlotFull
func EquipmentIsSlotFull(slotType *C.char) C.bool {
	if em == nil {
		return false
	}
	return C.bool(em.IsSlotFull(C.GoString(slotType)))
}

//export EquipmentIsItemEquipped
func EquipmentIsItemEquipped(slotType *C.char, itemID *C.char) C.bool {
	if em == nil {
		return false
	}
	return C.bool(em.IsItemEquipped(C.GoString(slotType), C.GoString(itemID)))
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
func EquipmentNextEquippedItem() *C.char {
	if em == nil {
		return C.CString("")
	}
	item := em.NextEquippedItem()
	if item == "" {
		return C.CString("")
	}
	return C.CString(item)
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
func EquipmentClearSlot(slotType *C.char) C.bool {
	if em == nil {
		return false
	}
	return C.bool(em.ClearSlot(C.GoString(slotType)))
}

//export EquipmentGetSlotAvailability
func EquipmentGetSlotAvailability(slotType *C.char) C.int {
	if em == nil {
		return 0
	}
	return C.int(em.GetSlotAvailability(C.GoString(slotType)))
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

// ---- Int ----

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
func ZoneConfig_GetMaxNPC( path *C.char, zoneID *C.char) C.int {
	id := C.GoString(zoneID)
	p := C.GoString(path)
	return C.int(zoneconfig.GetManager(p).GetMaxNPC(id))
}

//export ZoneConfig_GetSpawnChance
func ZoneConfig_GetSpawnChance(path *C.char, zoneID *C.char) C.float {
	id := C.GoString(zoneID)
	p := C.GoString(path)
	return C.float(zoneconfig.GetManager(p).GetSpawnChance(id))
}

//export ZoneConfig_GetRandomNPCType
func ZoneConfig_GetRandomNPCType(path *C.char, zoneID *C.char) *C.char {
	id := C.GoString(zoneID)
	p := C.GoString(path)
	npc := zoneconfig.GetManager(p).GetRandomNPCType(id)
	return C.CString(npc)
}

//export ZoneConfig_Reload
func ZoneConfig_Reload(path *C.char) C.int {
	p := C.GoString(path)
	err := zoneconfig.GetManager(p).Load()
	return C.int(err)
}

//export Voronoi_Init
func Voronoi_Init(width C.int, height C.int, numZones C.int, seed C.longlong) {
	voronoi.Init(int(width), int(height), int(numZones), int64(seed))
}

//export Voronoi_ZoneAt
func Voronoi_ZoneAt(x C.int, y C.int) C.int {
	return C.int(voronoi.ZoneAt(int(x), int(y)))
}

//export Voronoi_GetRandomInRadius
func Voronoi_GetRandomInRadius(x C.int, y C.int, zoneId C.int, radius C.int) C.Coord2d {
    px, py, ok := voronoi.RandomPositionInRadius(int(x), int(y), int(zoneId), int(radius));
	if !ok {
		return C.MakeCoord(C.int(-1), C.int(-1))
	}
    return C.MakeCoord(C.int(x), C.int(y))
}

//export Crafting_Register
func Crafting_Register(name *C.char, path *C.char) C.int{
	n := C.GoString(name)
	p := C.GoString(path)
	
	return C.int(crafting.Register(n, p))
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


func main() {}
