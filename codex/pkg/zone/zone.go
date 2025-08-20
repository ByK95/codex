// pkg/zone/zone.go
package zone

import (
	"fmt"
	"sync"
)

// Zone represents a single zone's threat
type Zone struct {
	ID     int32
	Threat float32
}

// ZoneManager is a singleton that tracks all zones
type ZoneManager struct {
	mu        sync.Mutex
	zones     map[int32]*Zone
	mapFactor float32
}

var manager *ZoneManager
var once sync.Once

// GetManager returns the singleton instance
func GetManager() *ZoneManager {
	once.Do(func() {
		manager = &ZoneManager{
			zones:     make(map[int32]*Zone),
			mapFactor: 1.0,
		}
	})
	return manager
}

// RegisterZone adds a new zone to the manager
func (m *ZoneManager) RegisterZone(z *Zone) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.zones[z.ID] = z
}

// IncreaseThreat adds threat to a specific zone without decaying others
func (m *ZoneManager) IncreaseThreat(zoneID int32, amount float32) float32 {
	m.mu.Lock()
	defer m.mu.Unlock()

	z, ok := m.zones[zoneID]
	if !ok {
		z = &Zone{ID: zoneID, Threat: 0}
		m.zones[zoneID] = z
	}

	z.Threat += amount
	return z.Threat
}

// TimedThreat increases current zone threat and decays all others
func (m *ZoneManager) TimedThreat(currentID int32, amount float32) float32 {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure current zone exists
	z, ok := m.zones[currentID]
	if !ok {
		z = &Zone{ID: currentID, Threat: 0}
		m.zones[currentID] = z
	}

	// Increase current zone threat
	z.Threat += amount

	// Decay other zones
	for id, other := range m.zones {
		if id == currentID {
			continue
		}
		other.Threat -= amount / m.mapFactor
		if other.Threat < 0 {
			other.Threat = 0
		}
	}

	return z.Threat
}

// AdvanceMap increases map factor (slows decay on new maps)
func (m *ZoneManager) AdvanceMap() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mapFactor += 1.0
}

// GetZoneThreat returns current threat of a zone
func (m *ZoneManager) GetZoneThreat(zoneID int32) float32 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if z, ok := m.zones[zoneID]; ok {
		return z.Threat
	}
	return 0
}

// Reset clears all zones and resets map factor
func (m *ZoneManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.zones = make(map[int32]*Zone)
	m.mapFactor = 1.0
}

// String prints all zone threats
func (m *ZoneManager) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := "Zones:\n"
	for id, z := range m.zones {
		s += fmt.Sprintf("  Zone[%d] Threat=%.2f\n", id, z.Threat)
	}
	return s
}
