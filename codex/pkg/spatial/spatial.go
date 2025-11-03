package spatial

import (
	"math"
	"sync"
)

type LootData struct {
    ID int
    X  float64
    Y  float64
}

type SpatialGrid struct {
    CellSize float64
    Width    int
    Height   int
    Cells    [][]*LootData
}


// Global spatial grids keyed by type ID or name
var (
    SpatialGrids = make(map[int]*SpatialGrid)
    gridsMu      sync.RWMutex
)

// Register a new spatial grid with an integer key
func RegisterGrid(id int, width, height int, cellSize float64) {
    gridsMu.Lock()
    defer gridsMu.Unlock()
    SpatialGrids[id] = NewSpatialGrid(width, height, cellSize)
}

// Get a grid by ID
func GetGrid(id int) *SpatialGrid {
    gridsMu.RLock()
    defer gridsMu.RUnlock()
    return SpatialGrids[id]
}

// Remove a grid
func RemoveGrid(id int) {
    gridsMu.Lock()
    defer gridsMu.Unlock()
    delete(SpatialGrids, id)
}


// NewSpatialGrid creates a new 2D grid
func NewSpatialGrid(width, height int, cellSize float64) *SpatialGrid {
	cells := make([][]*LootData, width*height)
	for i := range cells {
		cells[i] = []*LootData{}
	}
	return &SpatialGrid{
		CellSize: cellSize,
		Width:    width,
		Height:   height,
		Cells:    cells,
	}
}

// IndexFromCoords computes the grid cell index from world coordinates
func (g *SpatialGrid) IndexFromCoords(x, y float64) int {
	cellX := int(x / g.CellSize)
	cellY := int(y / g.CellSize)

	if cellX < 0 {
		cellX = 0
	} else if cellX >= g.Width {
		cellX = g.Width - 1
	}
	if cellY < 0 {
		cellY = 0
	} else if cellY >= g.Height {
		cellY = g.Height - 1
	}

	return cellX + cellY*g.Width
}

func (g *SpatialGrid) InsertSpatial(id int, x, y float64) {
	// Clamp positions to grid bounds
	if x < 0 {
		x = 0
	} else if x >= float64(g.Width)*g.CellSize {
		x = float64(g.Width)*g.CellSize - 0.001
	}
	if y < 0 {
		y = 0
	} else if y >= float64(g.Height)*g.CellSize {
		y = float64(g.Height)*g.CellSize - 0.001
	}

	loot := &LootData{
		ID: id,
		X:  x,
		Y:  y,
	}
	idx := g.IndexFromCoords(x, y)
	g.Cells[idx] = append(g.Cells[idx], loot)
}


// RemoveSpatial removes a loot item by its ID and coordinates
func (g *SpatialGrid) RemoveSpatial(id int, x, y float64) {
	idx := g.IndexFromCoords(x, y)
	slice := g.Cells[idx]
	for i, v := range slice {
		if v.ID == id {
			g.Cells[idx] = append(slice[:i], slice[i+1:]...)
			break
		}
	}
}

// ClosestSpatial returns the ID of the closest loot within radius, or -1 if none found
func (g *SpatialGrid) ClosestSpatial(x, y, radius float64, neighborCells int) int {
	minX := int(x/g.CellSize) - neighborCells
	maxX := int(x/g.CellSize) + neighborCells
	minY := int(y/g.CellSize) - neighborCells
	maxY := int(y/g.CellSize) + neighborCells

	if minX < 0 {
		minX = 0
	}
	if maxX >= g.Width {
		maxX = g.Width - 1
	}
	if minY < 0 {
		minY = 0
	}
	if maxY >= g.Height {
		maxY = g.Height - 1
	}

	radiusSq := radius * radius
	closestID := -1
	minDistSq := math.MaxFloat64

	for cx := minX; cx <= maxX; cx++ {
		for cy := minY; cy <= maxY; cy++ {
			idx := cx + cy*g.Width
			for _, loot := range g.Cells[idx] {
				dx := loot.X - x
				dy := loot.Y - y
				distSq := dx*dx + dy*dy
				if distSq <= radiusSq && distSq < minDistSq {
					minDistSq = distSq
					closestID = loot.ID
				}
			}
		}
	}

	return closestID
}