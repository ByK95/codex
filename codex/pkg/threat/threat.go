// pkg/threat/threat.go
package threat

import (
	"sync"
)

// Threat represents a single zone's threat
type Threat struct {
	ZoneID int32
	Value  float32
}

// ThreatManager is a singleton that tracks all zones
type ThreatManager struct {
	mu        sync.Mutex
	zones     map[int32]*Threat
	mapFactor float32
}

var manager *ThreatManager
var once sync.Once

// GetManager returns the singleton instance
func GetManager() *ThreatManager {
	once.Do(func() {
		manager = &ThreatManager{
			zones:     make(map[int32]*Threat),
			mapFactor: 1.0,
		}
	})
	return manager
}

// RegisterZone adds a new zone to the manager
func (m *ThreatManager) RegisterZone(t *Threat) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.zones[t.ZoneID] = t
}

// IncreaseThreat adds threat to a specific zone without decaying others
func (m *ThreatManager) IncreaseThreat(zoneID int32, amount float32) float32 {
	m.mu.Lock()
	defer m.mu.Unlock()

	z, ok := m.zones[zoneID]
	if !ok {
		z = &Threat{ZoneID: zoneID, Value: 0}
		m.zones[zoneID] = z
	}

	z.Value += amount
	return z.Value
}

// TimedThreat increases current zone threat and decays all others
func (m *ThreatManager) TimedThreat(currentID int32, amount float32) float32 {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure current zone exists
	z, ok := m.zones[currentID]
	if !ok {
		z = &Threat{ZoneID: currentID, Value: 0}
		m.zones[currentID] = z
	}

	// Increase current zone threat
	z.Value += amount

	// Decay other zones
	for id, other := range m.zones {
		if id == currentID {
			continue
		}
		other.Value -= amount / m.mapFactor
		if other.Value < 0 {
			other.Value = 0
		}
	}

	return z.Value
}

// AdvanceMap increases map factor (slows decay on new maps)
func (m *ThreatManager) AdvanceMap() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mapFactor += 1.0
}

// GetZoneThreat returns current threat of a zone
func (m *ThreatManager) GetZoneThreat(zoneID int32) float32 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if z, ok := m.zones[zoneID]; ok {
		return z.Value
	}
	return 0
}

// Reset clears all zones and resets map factor
func (m *ThreatManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.zones = make(map[int32]*Threat)
	m.mapFactor = 1.0
}
