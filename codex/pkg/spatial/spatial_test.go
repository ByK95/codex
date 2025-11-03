package spatial

import (
	"testing"
)

func TestInsertAndClosestSpatial(t *testing.T) {
	grid := NewSpatialGrid(10, 10, 10)

	// Insert some loot
	grid.InsertSpatial(1, 5, 5)
	grid.InsertSpatial(2, 15, 15)
	grid.InsertSpatial(3, 25, 25)

	// Closest within radius
	id := grid.ClosestSpatial(6, 6, 5, 0)
	if id != 1 {
		t.Errorf("Expected closest ID 1, got %d", id)
	}

	id = grid.ClosestSpatial(16, 14, 5, 0)
	if id != 2 {
		t.Errorf("Expected closest ID 2, got %d", id)
	}

	// No loot in radius
	id = grid.ClosestSpatial(0, 0, 2, 0)
	if id != -1 {
		t.Errorf("Expected -1 for no loot, got %d", id)
	}

	// Test neighborCells
	id = grid.ClosestSpatial(14, 14, 5, 1)
	if id != 2 {
		t.Errorf("Expected closest ID 2 with neighborCells=1, got %d", id)
	}
}

func TestRemoveSpatial(t *testing.T) {
	grid := NewSpatialGrid(10, 10, 10)

	grid.InsertSpatial(1, 5, 5)
	grid.InsertSpatial(2, 15, 15)

	// Remove loot
	grid.RemoveSpatial(1, 5, 5)

	id := grid.ClosestSpatial(5, 5, 5, 0)
	if id != -1 {
		t.Errorf("Expected -1 after removal, got %d", id)
	}

	// Ensure other loot still exists
	id = grid.ClosestSpatial(15, 15, 5, 0)
	if id != 2 {
		t.Errorf("Expected ID 2 to still exist, got %d", id)
	}
}

func TestEdgeCases(t *testing.T) {
	grid := NewSpatialGrid(10, 10, 10)

	// Insert loot outside the grid bounds (should clamp)
	grid.InsertSpatial(1, -5, -5)
	grid.InsertSpatial(2, 150, 150)

	id := grid.ClosestSpatial(0, 0, 20, 0)
	if id != 1 {
		t.Errorf("Expected ID 1 clamped to grid, got %d", id)
	}

	id = grid.ClosestSpatial(95, 95, 20, 0)
	if id != 2 {
		t.Errorf("Expected ID 2 clamped to grid, got %d", id)
	}

	// Insert multiple loot in same cell
	grid.InsertSpatial(3, 5, 5)
	grid.InsertSpatial(4, 5, 5)

	id = grid.ClosestSpatial(5, 5, 1, 0)
	if id != 3 && id != 4 {
		t.Errorf("Expected either ID 3 or 4, got %d", id)
	}
}
