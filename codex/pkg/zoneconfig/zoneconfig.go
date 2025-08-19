package zoneconfig

import (
	"encoding/json"
	"math/rand"
	"os"
	"sync"
)

type Zone struct {
	ZoneID      string   `json:"zone_id"`
	NPCs        []string `json:"npcs"`
	SpawnChance float32  `json:"spawn_chance"`
	MaxCount    int32    `json:"max_count"`
}

type ZoneConfigManager struct {
	mu    sync.RWMutex
	zones map[string]*Zone
	path  string
}

var manager *ZoneConfigManager
var once sync.Once

func GetManager(path string) *ZoneConfigManager {
	once.Do(func() {
		manager = &ZoneConfigManager{
			zones: make(map[string]*Zone),
			path:  path,
		}
		_ = manager.Load()
	})
	return manager
}

func (m *ZoneConfigManager) Load() int32 {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := os.ReadFile(m.path)
	if err != nil {
		return int32(-1)
	}

	var zones []*Zone
	if err := json.Unmarshal(data, &zones); err != nil {
		return int32(-2)
	}

	m.zones = make(map[string]*Zone)
	for _, z := range zones {
		m.zones[z.ZoneID] = z
	}
	return int32(1)
}

func (m *ZoneConfigManager) GetMaxNPC(zoneID string) int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if z, ok := m.zones[zoneID]; ok {
		return z.MaxCount
	}
	return 0
}

func (m *ZoneConfigManager) GetSpawnChance(zoneID string) float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if z, ok := m.zones[zoneID]; ok {
		return z.SpawnChance
	}
	return 0
}

func (m *ZoneConfigManager) GetRandomNPCType(zoneID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	z, ok := m.zones[zoneID]
	if !ok || len(z.NPCs) == 0 {
		return ""
	}

	index := rand.Intn(len(z.NPCs))
	return z.NPCs[index]
}
