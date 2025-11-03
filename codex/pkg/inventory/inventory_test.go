package inventory

import (
	"testing"
	"fmt"
	"github.com/stretchr/testify/assert"
)

func TestTakeOneFromSlot(t *testing.T) {
	draggedSlot := DraggedSlot{Empty: true}
	inventory := NewInventory(3)
	item := Item{ID: 1, Quantity: 10, Stackable: true, MaxStackSize: 20}
	inventory.Slots[0] = &item
	inventory.itemCounts[item.ID] = item.Quantity

	fmt.Println(inventory.Slots[0].Quantity)
	// Take one when dragged empty
	assert.True(t, inventory.TakeOneFromSlot(&draggedSlot, 0), "Should take one from slot when empty dragged")
	assert.False(t, draggedSlot.Empty, "draggedSlot should not be empty after taking one")
	assert.Equal(t, 1, draggedSlot.Item.Quantity, "draggedSlot quantity should be 1")
	assert.Equal(t, 9, inventory.Slots[0].Quantity, "Slot quantity should decrease to 9")

	// Take one when dragged has same item and can stack
	assert.True(t, inventory.TakeOneFromSlot(&draggedSlot, 0), "Should stack onto dragged slot")
	assert.Equal(t, 2, draggedSlot.Item.Quantity, "draggedSlot quantity should increase to 2")
	assert.Equal(t, 8, inventory.Slots[0].Quantity, "Slot quantity should decrease to 8")

	// Fill dragged slot to max stack
	draggedSlot.Item.Quantity = draggedSlot.Item.MaxStackSize
	assert.False(t, inventory.TakeOneFromSlot(&draggedSlot, 0), "Should not take when dragged slot is full")

	inventory.Slots[0].Quantity = 1
	draggedSlot.Item.Quantity = draggedSlot.Item.MaxStackSize - 1
	assert.True(t, inventory.TakeOneFromSlot(&draggedSlot, 0), "Should take last item when dragged")
	assert.Nil(t, inventory.Slots[0], "When all selected item should be")

	// Different item in slot while dragged slot not empty
	inventory.Slots[1] = &Item{ID: 2, Quantity: 5, Stackable: true, MaxStackSize: 10}
	assert.False(t, inventory.TakeOneFromSlot(&draggedSlot,1), "Should not take when dragged slot has different item")

	// Non-stackable item in slot
	inventory.Slots[2] = &Item{ID: 3, Quantity: 1, Stackable: false, MaxStackSize: 1}
	draggedSlot.Empty = true
	assert.False(t, inventory.TakeOneFromSlot(&draggedSlot, 2), "Should not take when item is not stackable")

	// Out of bounds index
	assert.False(t, inventory.TakeOneFromSlot(&draggedSlot, -1), "Should not take for invalid slot index")
	assert.False(t, inventory.TakeOneFromSlot(&draggedSlot, len(inventory.Slots)), "Should not take for invalid slot index")
}

func TestDropToSlot(t *testing.T) {
	inventoryID := NewInventoryInstance(3)
	inventory := GetInventory(inventoryID)

	// Setup: draggedSlot with 5 of item ID=1
	draggedSlot := DraggedSlot{
		Item:  &Item{ID: 1, Quantity: 5, Stackable: true, MaxStackSize: 10},
		Empty: false,
		OriginIdx: 1,
	}

	// Case 1: drop into empty slot
	assert.True(t, inventory.DropToSlot(&draggedSlot, 0))
	assert.True(t, draggedSlot.Empty)
	assert.Nil(t, draggedSlot.Item)
	slot0 := inventory.Slots[0]
	assert.NotNil(t, slot0)
	assert.Equal(t, 5, slot0.Quantity)
	assert.Equal(t, 5, inventory.itemCounts[1])

	// Reset draggedSlot for next test
	draggedSlot = DraggedSlot{
		Item:  &Item{ID: 1, Quantity: 6, Stackable: true, MaxStackSize: 10},
		Empty: false,
	}

	// Setup slot 1 with same item but quantity 5 (partial stack)
	inventory.Slots[1] = &Item{ID: 1, Quantity: 5, Stackable: true, MaxStackSize: 10}
	inventory.itemCounts[1] = 5

	// Case 2: drop on same stackable item, total under max stack size (5 + 6 = 11 > 10)
	assert.True(t, inventory.DropToSlot(&draggedSlot, 1))
	slot1 := inventory.Slots[1]
	assert.NotNil(t, slot1)
	assert.Equal(t, 10, slot1.Quantity)                  // maxed stack
	assert.False(t, draggedSlot.Empty)                   // leftover dragged
	assert.Equal(t, 1, draggedSlot.Item.Quantity)        // leftover = 11 - 10 = 1
	assert.Equal(t, 10, inventory.itemCounts[1])               // updated count

	// Case 3: drop on different item should swap
	draggedSlot = DraggedSlot{
		Item:  &Item{ID: 2, Quantity: 3, Stackable: true, MaxStackSize: 5},
		Empty: false,
		OriginIdx: 1,
		OriginInvID: inventoryID,
	}
	inventory.Slots[2] = &Item{ID: 3, Quantity: 4, Stackable: true, MaxStackSize: 10}
	inventory.itemCounts[2] = 0
	inventory.itemCounts[3] = 4

	assert.True(t, inventory.DropToSlot(&draggedSlot, 2))
	slot2 := inventory.Slots[2]
	assert.NotNil(t, slot2)
	assert.Equal(t, 3, slot2.Quantity)      // now slot has dragged item
	assert.Equal(t, 3, inventory.Slots[1].ID) // dragged now holds old slot item id 3
	assert.Equal(t, 4, inventory.Slots[1].Quantity)
	assert.Equal(t, 3, inventory.itemCounts[2])
	assert.Equal(t, 4, inventory.itemCounts[3])
}

func TestPickUpFromSlot(t *testing.T) {
	inventory := NewInventory(3)

	// Setup slot 0 with an item
	item := &Item{ID: 1, Quantity: 5, Stackable: true, MaxStackSize: 10}
	inventory.Slots[0] = item
	inventory.itemCounts[item.ID] = item.Quantity
	draggedSlot := DraggedSlot{Empty: true}

	// Pick up when draggedSlot empty: should succeed and clear slot
	assert.True(t, inventory.PickUpFromSlot(&draggedSlot, 0))
	assert.False(t, draggedSlot.Empty)
	assert.Equal(t, item, draggedSlot.Item)
	assert.Nil(t, inventory.Slots[0])
	assert.Equal(t, 0, inventory.itemCounts[item.ID])

	// Pick up from empty slot: fail
	assert.False(t, inventory.PickUpFromSlot(&draggedSlot, 0))

	// Setup slot 1 and draggedSlot with different items, test swap
	item2 := &Item{ID: 2, Quantity: 3, Stackable: true, MaxStackSize: 10}
	inventory.Slots[1] = item2
	// draggedSlot already holding item from previous pickup
	assert.True(t, inventory.PickUpFromSlot(&draggedSlot, 1))
	assert.Equal(t, item2, draggedSlot.Item)
	assert.NotNil(t, inventory.Slots[1])
	assert.Equal(t, 5, inventory.Slots[1].Quantity) // previous dragged item swapped here
}


func TestAddItem(t *testing.T) {
	inventory := NewInventory(3)

	stackableID := 1
	stackableMax := 5

	// Add stackable items fewer than max stack size
	assert.True(t, inventory.AddItem(stackableID, true, stackableMax, 3))
	assert.NotNil(t, inventory.Slots[0])
	assert.Equal(t, 3, inventory.Slots[0].Quantity)
	assert.Equal(t, 3, inventory.itemCounts[stackableID])
	assert.Contains(t, inventory.partialStacks[stackableID], 0)

	// Add more to fill stack
	assert.True(t, inventory.AddItem(stackableID, true, stackableMax, 2))
	assert.Equal(t, 5, inventory.Slots[0].Quantity)
	assert.Equal(t, 5, inventory.itemCounts[stackableID])
	assert.NotContains(t, inventory.partialStacks[stackableID], 0)

	// Add exceeding max stack size to create new stack
	assert.True(t, inventory.AddItem(stackableID, true, stackableMax, 6))
	assert.Equal(t, 5, inventory.Slots[0].Quantity)
	assert.NotNil(t, inventory.Slots[1])
	assert.Equal(t, 5, inventory.Slots[1].Quantity)
	assert.Equal(t, 11, inventory.itemCounts[stackableID])
}

func TestAddNotStackable(t *testing.T) {
	inventory := NewInventory(3)

	nonStackableID := 2
	nonStackableMax := 1

	// Add non-stackable items, each occupies one slot
	assert.True(t, inventory.AddItem(nonStackableID, false, nonStackableMax, 2))
	assert.NotNil(t, inventory.Slots[0])
	assert.Equal(t, 1, inventory.Slots[0].Quantity)
	assert.Equal(t, 2, inventory.itemCounts[nonStackableID])

	// Adding more non-stackable than slots available fails partially
	assert.False(t, inventory.AddItem(nonStackableID, false, nonStackableMax, 2)) // only one slot left, so only one added
	assert.Equal(t, 3, len(inventory.Slots))
}

func TestRemainingCapacity(t *testing.T) {
	inv := NewInventory(5)
	stackableID := 1
	maxStack := 10

	// Empty inventory: all slots available
	assert.Equal(t, 50, inv.RemainingCapacity(stackableID, true, maxStack))

	// Fill one slot partially
	inv.Slots[0] = &Item{ID: stackableID, Quantity: 6, Stackable: true, MaxStackSize: maxStack}
	inv.partialStacks[stackableID] = []int{0}
	inv.itemCounts[stackableID] = 6
	assert.Equal(t, 44, inv.RemainingCapacity(stackableID, true, maxStack)) // 4 left in stack + 4 empty slots * 10 = 44

	// Fill another slot completely
	inv.Slots[1] = &Item{ID: stackableID, Quantity: 10, Stackable: true, MaxStackSize: maxStack}
	delete(inv.partialStacks, stackableID) // slot 1 is full
	inv.partialStacks[stackableID] = []int{0}
	assert.Equal(t, 34, inv.RemainingCapacity(stackableID, true, maxStack)) // 4 + (3 * 10)

	// Non-stackable case: each slot can take only 1 item
	inv2 := NewInventory(3)
	assert.Equal(t, 3, inv2.RemainingCapacity(2, false, 1))
	inv2.Slots[0] = &Item{ID: 2, Quantity: 1, Stackable: false, MaxStackSize: 1}
	assert.Equal(t, 2, inv2.RemainingCapacity(2, false, 1))
}

func TestPickUpAndDropSwap(t *testing.T) {
	invID := NewInventoryInstance(2)
	inv := GetInventory(invID)

	// slot 0: item A (id=1)
	inv.AddItem(1, false, 1, 1)
	// slot 1: item B (id=2)
	inv.AddItem(2, false, 1, 1)

	dragged := &DraggedSlot{Empty: true}

	// Pick up from slot 0 (item A)
	ok := inv.PickUpFromSlot(dragged, 0)
	if !ok || dragged.Empty {
		t.Fatalf("expected to pick up item A from slot 0")
	}

	// Drop it on slot 1 (item B)
	ok = inv.DropToSlot(dragged, 1)
	if !ok {
		t.Fatalf("expected to drop/swap with slot 1")
	}

	// After swap:
	// slot 0 should have item B
	if inv.Slots[0] == nil || inv.Slots[0].ID != 2 {
		t.Errorf("expected slot 0 to have item ID 2, got %+v", inv.Slots[0])
	}

	// slot 1 should have item A
	if inv.Slots[1] == nil || inv.Slots[1].ID != 1 {
		t.Errorf("expected slot 1 to have item ID 1, got %+v", inv.Slots[1])
	}

	// draggedSlot should now be empty
	if !dragged.Empty {
		t.Errorf("expected dragged slot to be empty after drop")
	}
}

func TestDropIntoPartialStack(t *testing.T) {
	invID := NewInventoryInstance(2)
	inv := GetInventory(invID)

	// slot 0: item A (id=1), quantity 5, max stack 10
	inv.AddItem(1, true, 10, 5)

	// dragged slot: item A (id=1), quantity 4
	dragged := &DraggedSlot{
		Item:  &Item{ID: 1, Quantity: 4, Stackable: true, MaxStackSize: 10},
		Empty: false,
	}

	// Drop dragged stack onto slot 0
	ok := inv.DropToSlot(dragged, 0)
	if !ok {
		t.Fatalf("expected to drop onto partial stack")
	}

	slot0 := inv.Slots[0]
	if slot0 == nil || slot0.ID != 1 {
		t.Fatalf("expected slot 0 to still hold item A")
	}

	// Total quantity should be 9
	if slot0.Quantity != 9 {
		t.Errorf("expected slot 0 quantity 9, got %d", slot0.Quantity)
	}

	// Dragged slot should now be empty (all items merged)
	if !dragged.Empty {
		t.Errorf("expected dragged slot to be empty after merge")
	}

	// itemCounts should be updated correctly
	if inv.itemCounts[1] != 9 {
		t.Errorf("expected itemCounts[1] = 9, got %d", inv.itemCounts[1])
	}
}

func TestFullStackSwap(t *testing.T) {
	invID := NewInventoryInstance(2)
	inv := GetInventory(invID)

	// Slot 0: item A, full stack (10/10)
	inv.Slots[0] = &Item{ID: 1, Quantity: 10, Stackable: true, MaxStackSize: 10}
	inv.itemCounts[1] = 10

	// Slot 1: item A, also full stack (10/10)
	inv.Slots[1] = &Item{ID: 1, Quantity: 10, Stackable: true, MaxStackSize: 10}
	inv.itemCounts[1] += 10

	// DraggedSlot: item A, full stack (10/10)
	dragged := &DraggedSlot{
		Item:       &Item{ID: 1, Quantity: 10, Stackable: true, MaxStackSize: 10},
		Empty:      false,
		OriginIdx:  0,
		OriginInvID: invID,
	}

	// Drop dragged full stack on another full stack of same item
	ok := inv.DropToSlot(dragged, 1)
	if !ok {
		t.Fatalf("expected DropToSlot to succeed for full stack swap")
	}

	// Dragged slot should now be empty
	if !dragged.Empty || dragged.Item != nil {
		t.Errorf("expected dragged slot to be empty after full stack swap, got %+v", dragged.Item)
	}

	// The items should be swapped in place (still same ID but technically swapped instances)
	slot0 := inv.Slots[0]
	slot1 := inv.Slots[1]

	if slot0 == nil || slot1 == nil {
		t.Fatalf("expected both slots to be non-nil")
	}
	if slot0 == slot1 {
		t.Errorf("expected slots 0 and 1 to be different instances after swap")
	}
	if slot0.ID != 1 || slot1.ID != 1 {
		t.Errorf("expected both slots to have item ID 1, got %v and %v", slot0.ID, slot1.ID)
	}
}

func TestDropIntoFullStackDifferentInventory(t *testing.T) {
	// Inventory A: origin inventory
	invAID := NewInventoryInstance(2)
	invA := GetInventory(invAID)
	invA.AddItem(1, true, 10, 10)

	// Inventory B: target inventory
	invBID := NewInventoryInstance(2)
	invB := GetInventory(invBID)
	invB.AddItem(2, true, 10, 10) // full stack

	// Drag item from invA slot 0
	dragged := &DraggedSlot{Empty: true}
	ok := invA.PickUpFromSlot(dragged, 0)
	if !ok {
		t.Fatalf("failed to pick up from invA")
	}

	// Drop into invB slot 0 (full stack)
	ok = invB.DropToSlot(dragged, 0)
	if !ok {
		t.Fatalf("DropToSlot failed unexpectedly")
	}

	// Inspect results
	slotB := invB.Slots[0]
	if slotB.Quantity != 10 || slotB.ID != 1 {
		t.Errorf("expected invB slot0 quantity 10, got %d ID: %d", slotB.Quantity, slotB.ID)
	}

	if invA.Slots[0].Quantity != 10 || invA.Slots[0].ID != 2 {
		t.Errorf("expected invA slot0 quantity 10, got %d ID: %d", invA.Slots[0].Quantity, invA.Slots[0].ID)
	}
	
}

