package loot

import (
	"math/rand"
	"sync"
	"time"
)

var (
	pityCounts = make(map[int32]int32) // itemID -> pity counter
	mu         sync.Mutex
)

type LootRow struct {
    ID     int32
    Chance float32
    Pity   int32
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ResetPity clears all pity counters
func ResetPity() {
	mu.Lock()
	defer mu.Unlock()
	pityCounts = make(map[int32]int32)
}

// RollLoot rolls for each item based on chance and pity
// Params must have the same length: ids[i], chances[i], pityThresholds[i]
func RollLoot(items []LootRow) map[int32]int32 {
	mu.Lock()
	defer mu.Unlock()

	results := make(map[int32]int32)

	for i := 0; i < len(items); i++ {
		id := items[i].ID
		chance := items[i].Chance
		pity := items[i].Pity

		// Guaranteed drop
		if chance >= 1.0 {
			results[id]++
			pityCounts[id] = 0
			continue
		}

		// Increment pity counter
		pityCounts[id]++

		// Check pity threshold
		if pity > 0 && pityCounts[id] > pity {
			results[id]++
			pityCounts[id] = 0
			continue
		}

		// Roll chance
		if rand.Float32() < chance {
			results[id]++
			pityCounts[id] = 0
		}
	}

	// return key for caching in rollResults store in roll results 
	return results
}