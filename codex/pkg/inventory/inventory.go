package inventory

import "codex/pkg/iterator"

type Item struct {
	ID           int
	Quantity     int
	Stackable    bool
	MaxStackSize int
}

type Inventory struct {
	ID    int
	Slots []*Item

	// itemID → total quantity in inventory (O(1) count)
	itemCounts map[int]int

	// itemID → slice of slot indexes that contain that item with space for stacking
	partialStacks map[int][]int
}

type DraggedSlot struct {
	Item        *Item
	Empty       bool
	OriginIdx   int
	OriginInvID int
}

var (
	inventories = make(map[int]*Inventory)
	nextInvID   = 1
	draggedSlot = DraggedSlot{Empty: true}
	itemIter *iterator.Iterator[int]
)

// Creates a new inventory instance and returns its ID
func NewInventoryInstance(slotCount int) int {
	id := nextInvID
	nextInvID++
	inventories[id] = NewInventory(slotCount)
	inventories[id].ID = id
	return id
}

// Deletes a specific inventory instance by ID.
// Returns true if deleted, false if not found.
func DeleteInventoryInstance(id int) bool {
	if _, ok := inventories[id]; !ok {
		return false
	}
	delete(inventories, id)
	return true
}

// Clears all inventories and resets ID counter.
func ClearAllInventories() {
	inventories = make(map[int]*Inventory)
	nextInvID = 1
	ResetDraggedSlot()
}

// Returns an inventory by ID
func GetInventory(id int) *Inventory {
	return inventories[id]
}

// Returns a pointer to the global dragged slot
func GetDraggedSlot() *DraggedSlot {
	return &draggedSlot
}

func ResetDraggedSlot() {
	draggedSlot = DraggedSlot{Empty: true}
}

// Moves the dragged item back to its origin slot (if possible).
// Returns true if restored, false otherwise.
func RestoreDraggedItem(draggedSlot *DraggedSlot) bool {
	if draggedSlot.Empty || draggedSlot.Item == nil {
		return false
	}

	origin := GetInventory(draggedSlot.OriginInvID)
	if origin == nil {
		return false
	}

	idx := draggedSlot.OriginIdx
	if idx < 0 || idx >= len(origin.Slots) {
		return false
	}

	// If the original slot is free, restore it directly
	if origin.Slots[idx] == nil {
		origin.Slots[idx] = draggedSlot.Item
		origin.itemCounts[draggedSlot.Item.ID] += draggedSlot.Item.Quantity
		ResetDraggedSlot()
		return true
	}

	// If occupied, try to add item somewhere else
	ok := origin.AddItem(
		draggedSlot.Item.ID,
		draggedSlot.Item.Stackable,
		draggedSlot.Item.MaxStackSize,
		draggedSlot.Item.Quantity,
	)
	if ok {
		ResetDraggedSlot()
		return true
	}

	// Couldn't restore — leave draggedSlot unchanged
	return false
}

func CancelDraggedSlot() bool {
	if draggedSlot.Empty || draggedSlot.Item == nil {
		return false
	}
	origin := GetInventory(draggedSlot.OriginInvID)
	if origin == nil {
		return false
	}
	if draggedSlot.OriginIdx < 0 || draggedSlot.OriginIdx >= len(origin.Slots) {
		return false
	}

	// If origin slot already occupied, try to add elsewhere
	if origin.Slots[draggedSlot.OriginIdx] != nil {
		ok := origin.AddItem(
			draggedSlot.Item.ID,
			draggedSlot.Item.Stackable,
			draggedSlot.Item.MaxStackSize,
			draggedSlot.Item.Quantity,
		)
		if !ok {
			return false
		}
	} else {
		origin.Slots[draggedSlot.OriginIdx] = draggedSlot.Item
		origin.itemCounts[draggedSlot.Item.ID] += draggedSlot.Item.Quantity
	}

	ResetDraggedSlot()
	return true
}

func NewInventory(slotCount int) *Inventory {
	return &Inventory{
		Slots:         make([]*Item, slotCount),
		itemCounts:    make(map[int]int),
		partialStacks: make(map[int][]int),
	}
}

// Adds item to inventory
func (inv *Inventory) AddItem(id int, stackable bool, maxStackSize int, qty int) bool {
	if stackable {
		// Fill partial stacks first
		slots, ok := inv.partialStacks[id]
		if ok {
			newSlots := slots[:0]
			for _, idx := range slots {
				slot := inv.Slots[idx]
				if slot == nil || slot.ID != id {
					continue // skip and remove invalid entry
				}
				space := slot.MaxStackSize - slot.Quantity
				if space <= 0 {
					continue
				}
				add := min(qty, space)
				slot.Quantity += add
				inv.itemCounts[id] += add
				qty -= add
				if slot.Quantity < slot.MaxStackSize {
					newSlots = append(newSlots, idx)
				}
				if qty == 0 {
					inv.partialStacks[id] = newSlots
					return true
				}
			}
			inv.partialStacks[id] = newSlots
		}
	}

	// Add new stacks in empty slots
	for i, slot := range inv.Slots {
		if slot == nil || slot.Quantity == 0 {
			add := qty
			if stackable {
				add = min(qty, maxStackSize)
			} else {
				//Non stackable always will be 1
				add = 1
			}
			newItem := &Item{
				ID:           id,
				Quantity:     add,
				Stackable:    stackable,
				MaxStackSize: maxStackSize,
			}
			inv.Slots[i] = newItem
			inv.itemCounts[id] += add
			qty -= add

			if stackable && add < maxStackSize {
				inv.partialStacks[id] = append(inv.partialStacks[id], i)
			}
			if qty == 0 {
				return true
			}
		}
	}
	return qty == 0
}

// Returns how many items of the given ID could still fit in this inventory.
// Considers existing partial stacks and empty slots.
func (inv *Inventory) RemainingCapacity(id int, stackable bool, maxStackSize int) int {
	totalCapacity := 0

	if stackable {
		// Fill remaining space in partial stacks first
		if slots, ok := inv.partialStacks[id]; ok {
			for _, idx := range slots {
				slot := inv.Slots[idx]
				if slot != nil && slot.Quantity < slot.MaxStackSize {
					totalCapacity += slot.MaxStackSize - slot.Quantity
				}
			}
		}
	}

	// Count all empty slots
	for _, slot := range inv.Slots {
		if slot == nil || slot.Quantity == 0 {
			if stackable {
				totalCapacity += maxStackSize
			} else {
				totalCapacity += 1
			}
		}
	}

	return totalCapacity
}

func (inv *Inventory) removePartialStack(itemID int, slotIdx int) {
	slots := inv.partialStacks[itemID]
	for i, idx := range slots {
		if idx == slotIdx {
			inv.partialStacks[itemID] = append(slots[:i], slots[i+1:]...)
			break
		}
	}
}

func (inv *Inventory) RemoveItem(id int, qty int) bool {
	for i, slot := range inv.Slots {
		if slot != nil && slot.ID == id {
			remove := min(qty, slot.Quantity)
			slot.Quantity -= remove
			inv.itemCounts[id] -= remove
			qty -= remove

			if slot.Quantity == 0 {
				inv.removePartialStack(id, i)
				inv.Slots[i] = nil
			} else if slot.Stackable && slot.Quantity < slot.MaxStackSize {
				inv.addPartialStack(id, i)
			}

			if qty == 0 {
				return true
			}
		}
	}
	return qty == 0
}

func (inv *Inventory) addPartialStack(itemID int, slotIdx int) {
	slots := inv.partialStacks[itemID]
	for _, idx := range slots {
		if idx == slotIdx {
			return // already present
		}
	}
	inv.partialStacks[itemID] = append(inv.partialStacks[itemID], slotIdx)
}

func (inv *Inventory) CountItem(id int) int {
	return inv.itemCounts[id]
}

func (inv *Inventory) PickUpFromSlot(draggedSlot *DraggedSlot, slotIdx int) bool {
	if slotIdx < 0 || slotIdx >= len(inv.Slots) {
		return false
	}

	slot := inv.Slots[slotIdx]
	if slot == nil {
		return false
	}

	if !draggedSlot.Empty {
		inv.swapSlots(draggedSlot, slotIdx)
		return true
	}

	draggedSlot.Item = slot
	draggedSlot.Empty = false
	draggedSlot.OriginIdx = slotIdx
	draggedSlot.OriginInvID = inv.ID
	inv.Slots[slotIdx] = nil
	inv.itemCounts[draggedSlot.Item.ID] -= draggedSlot.Item.Quantity

	return true
}

func (inv *Inventory) DropToSlot(draggedSlot *DraggedSlot, targetIdx int) bool {
	if draggedSlot.Empty || targetIdx < 0 || targetIdx >= len(inv.Slots) {
		return false
	}

	target := inv.Slots[targetIdx]

	if target == nil || target.Quantity == 0 {
		// Empty target slot: move all dragged items there
		inv.Slots[targetIdx] = draggedSlot.Item
		inv.itemCounts[draggedSlot.Item.ID] += draggedSlot.Item.Quantity
		draggedSlot.Empty = true
		draggedSlot.Item = nil
		return true
	}

	fullStackCheck := target.ID == draggedSlot.Item.ID && target.MaxStackSize == target.Quantity && draggedSlot.Item.Quantity == draggedSlot.Item.MaxStackSize
	// If item IDs differ OR items are same but not stackable, swap
	if target.ID != draggedSlot.Item.ID || !target.Stackable || fullStackCheck {
		inv.swapSlots(draggedSlot, targetIdx)
		GetInventory(draggedSlot.OriginInvID).swapSlots(draggedSlot, draggedSlot.OriginIdx)
		draggedSlot.Empty = true
		draggedSlot.Item = nil
		draggedSlot.OriginIdx = -1
		return true
	}

	// Same item and stackable: add as much as possible to target stack, keep leftovers dragged
	totalQty := target.Quantity + draggedSlot.Item.Quantity
	if totalQty <= target.MaxStackSize {
		target.Quantity = totalQty
		inv.itemCounts[target.ID] += draggedSlot.Item.Quantity
		draggedSlot.Empty = true
		draggedSlot.Item = nil
	} else {
		toAdd := target.MaxStackSize - target.Quantity
		target.Quantity = target.MaxStackSize
		inv.itemCounts[target.ID] += toAdd
		draggedSlot.Item.Quantity -= toAdd
		// draggedSlot keeps leftover quantity
	}

	// Update partialStacks accordingly
	if target.Quantity == target.MaxStackSize {
		inv.removePartialStack(target.ID, targetIdx)
	} else {
		inv.addPartialStack(target.ID, targetIdx)
	}

	return true
}

func (inv *Inventory) swapSlots(draggedSlot *DraggedSlot, targetIdx int) {
	target := inv.Slots[targetIdx]

	// Update counts before swap
	if target != nil {
		inv.itemCounts[target.ID] -= target.Quantity
	}

	inv.Slots[targetIdx] = draggedSlot.Item
	if target != nil {
		oldTarget := *target
		draggedSlot.Item = &oldTarget
	}

	// Update counts after swap
	inv.itemCounts[inv.Slots[targetIdx].ID] += inv.Slots[targetIdx].Quantity
}

func (inv *Inventory) TakeOneFromSlot(draggedSlot *DraggedSlot, slotIdx int) bool {
	if slotIdx < 0 || slotIdx >= len(inv.Slots) {
		return false
	}
	slot := inv.Slots[slotIdx]
	if slot == nil || slot.Quantity == 0 {
		return false
	}

	if !slot.Stackable {
		return false
	}

	if draggedSlot.Empty {
		draggedSlot.Item = &Item{
			ID:           slot.ID,
			Quantity:     1,
			Stackable:    slot.Stackable,
			MaxStackSize: slot.MaxStackSize,
		}
		draggedSlot.Empty = false
		slot.Quantity--
		inv.itemCounts[slot.ID]--
		if slot.Quantity == 0 {
			inv.removePartialStack(slot.ID, slotIdx)
			inv.Slots[slotIdx] = nil
		}
		return true
	}

	if !draggedSlot.Empty && draggedSlot.Item.ID != slot.ID {
		return false
	}

	if !draggedSlot.Empty && draggedSlot.Item.ID == slot.ID {
		if draggedSlot.Item.Quantity >= draggedSlot.Item.MaxStackSize {
			// Can't exceed max stack size
			return false
		}
		draggedSlot.Item.Quantity++
		slot.Quantity--
		inv.itemCounts[slot.ID]--

		if slot.Quantity == 0 {
			inv.removePartialStack(slot.ID, slotIdx)
			inv.Slots[slotIdx] = nil
		} else if slot.Stackable && slot.Quantity < slot.MaxStackSize {
			inv.addPartialStack(slot.ID, slotIdx)
		}
		return true
	}

	return false
}

// InitItemIDIter initializes an iterator over all item IDs in the inventory.
// Returns the total number of unique item IDs.
func (inv *Inventory) InitItemIDIter() int {
	keys := make([]int, 0, len(inv.itemCounts))
	for id := range inv.itemCounts {
		keys = append(keys, id)
	}
	itemIter = iterator.NewIterator(keys)
	return len(keys)
}

// NextItemID returns the next item ID from the iterator.
// Returns -1 if iteration is complete or uninitialized.
func (inv *Inventory) NextItemID() int {
	if itemIter == nil {
		return -1
	}
	val, ok := itemIter.Next()
	if !ok {
		return -1
	}
	return val
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
