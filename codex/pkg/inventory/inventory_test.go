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
	inventory := NewInventory(3)

	// Setup: draggedSlot with 5 of item ID=1
	draggedSlot := DraggedSlot{
		Item:  &Item{ID: 1, Quantity: 5, Stackable: true, MaxStackSize: 10},
		Empty: false,
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
	}
	inventory.Slots[2] = &Item{ID: 3, Quantity: 4, Stackable: true, MaxStackSize: 10}
	inventory.itemCounts[2] = 0
	inventory.itemCounts[3] = 4

	assert.True(t, inventory.DropToSlot(&draggedSlot, 2))
	slot2 := inventory.Slots[2]
	assert.NotNil(t, slot2)
	assert.Equal(t, 3, slot2.Quantity)      // now slot has dragged item
	assert.Equal(t, 3, draggedSlot.Item.ID) // dragged now holds old slot item id 3
	assert.Equal(t, 4, draggedSlot.Item.Quantity)
	assert.Equal(t, 3, inventory.itemCounts[2])
	assert.Equal(t, 0, inventory.itemCounts[3])
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