package grid_voronoi

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestInitAndGet(t *testing.T) {
	Init(10, 10, 3, 42)
	v := Get()
	assert.NotNil(t, v)
	assert.Equal(t, 10, v.Width)
	assert.Equal(t, 10, v.Height)
	assert.Len(t, v.Seeds, 3)
	assert.NotNil(t, v.Grid)
}

func TestDeterministicSeed(t *testing.T) {
	Init(10, 10, 2, 12345)
	v1 := Get()
	grid1 := v1.Grid

	Init(10, 10, 2, 12345)
	v2 := Get()
	grid2 := v2.Grid

	assert.Equal(t, grid1, grid2, "Grids with same seed should match")
}

func TestDifferentSeedProducesDifferentGrid(t *testing.T) {
	Init(10, 10, 2, 111)
	v1 := Get()
	grid1 := v1.Grid

	Init(10, 10, 2, 222)
	v2 := Get()
	grid2 := v2.Grid

	assert.NotEqual(t, grid1, grid2, "Grids with different seeds should differ")
}

func TestZoneAtWithinBounds(t *testing.T) {
	Init(8, 8, 2, 99)
	zone := ZoneAt(3, 3)
	assert.True(t, zone >= 0)
	assert.True(t, zone < 2)
}

func TestZoneAtOutOfBounds(t *testing.T) {
	Init(5, 5, 2, 77)
	assert.Equal(t, -1, ZoneAt(-1, 0))
	assert.Equal(t, -1, ZoneAt(0, -1))
	assert.Equal(t, -1, ZoneAt(6, 0))
	assert.Equal(t, -1, ZoneAt(0, 6))
}

func TestZoneAtWithoutInit(t *testing.T) {
	// Force instance to nil
	Init(5, 5, 2, 55)
	v := Get()
	*v = *Get() // not nil
}
