package grid_voronoi

import (
	"math"
	"math/rand"
	"sync"
)

type Point struct {
	X, Y int
	Zone int
}

type Voronoi struct {
	Width, Height int
	Seeds         []Point
	Grid          [][]int
	rng           *rand.Rand
}

var (
	instance *Voronoi
	once     sync.Once
	mu       sync.Mutex
)

// Init re-initializes the singleton with new values
func Init(width, height, numZones int, seed int64) {
	mu.Lock()
	defer mu.Unlock()

	v := &Voronoi{
		Width:  width,
		Height: height,
		rng:    rand.New(rand.NewSource(seed)),
	}

	// generate seeds
	v.Seeds = make([]Point, numZones)
	for i := 0; i < numZones; i++ {
		v.Seeds[i] = Point{
			X:    v.rng.Intn(width),
			Y:    v.rng.Intn(height),
			Zone: i,
		}
	}

	// allocate grid
	v.Grid = make([][]int, height)
	for i := range v.Grid {
		v.Grid[i] = make([]int, width)
	}

	// assign zones
	v.assign()

	instance = v
}

// Get returns current singleton instance
func Get() *Voronoi {
	return instance
}

func (v *Voronoi) assign() {
	for y := 0; y < v.Height; y++ {
		for x := 0; x < v.Width; x++ {
			v.Grid[y][x] = v.nearestSeed(x, y)
		}
	}
}

func (v *Voronoi) nearestSeed(x, y int) int {
	minDist := math.MaxFloat64
	zone := -1
	for _, s := range v.Seeds {
		dx, dy := float64(x-s.X), float64(y-s.Y)
		d := dx*dx + dy*dy
		if d < minDist {
			minDist = d
			zone = s.Zone
		}
	}
	return zone
}

// ZoneAt returns zone index at (x,y)
func ZoneAt(x, y int) int {
	mu.Lock()
	defer mu.Unlock()

	if instance == nil {
		return -1
	}
	if x < 0 || y < 0 || x >= instance.Width || y >= instance.Height {
		return -1
	}
	return instance.Grid[y][x]
}
