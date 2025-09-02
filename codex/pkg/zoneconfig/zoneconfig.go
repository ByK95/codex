package zoneconfig

import (
	"codex/pkg/storage"
	"encoding/json"
	"fmt"
	"math/rand"
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
}

var manager *ZoneConfigManager
var once sync.Once

func init() {
    // Register load and save functions
    storage.SM().BindFuncs("zones", Load, nil)
}

func GetManager() *ZoneConfigManager {
	once.Do(func() {
		manager = &ZoneConfigManager{
			zones: make(map[string]*Zone),
		}
	})
	return manager
}

func Load(data json.RawMessage) error {
	var zones []*Zone
	if err := json.Unmarshal(data, &zones); err != nil {
		return fmt.Errorf("failed to unmarshal zones: %w", err)
	}

	GetManager().zones = make(map[string]*Zone)
	for _, z := range zones {
		GetManager().zones[z.ZoneID] = z
	}
	return nil
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
