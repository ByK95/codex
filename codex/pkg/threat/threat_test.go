// pkg/threat/threat_test.go
package threat

import (
	"testing"
)

func TestThreatManager_TimedThreat(t *testing.T) {
	manager := GetManager()
	manager.Reset()

	// Register threats
	manager.RegisterZone(&Threat{ZoneID: 1})
	manager.RegisterZone(&Threat{ZoneID: 2})
	manager.RegisterZone(&Threat{ZoneID: 3})

	// Increase threat for zone 1
	threat := manager.TimedThreat(1, 2.0)
	if threat != 2.0 {
		t.Errorf("Expected zone 1 threat 2.0, got %.2f", threat)
	}

	// Other zones should decay (amount / mapFactor = 2/1)
	if z2 := manager.GetZoneThreat(2); z2 != 0 {
		t.Errorf("Expected zone 2 threat 0 after decay, got %.2f", z2)
	}

	if z3 := manager.GetZoneThreat(3); z3 != 0 {
		t.Errorf("Expected zone 3 threat 0 after decay, got %.2f", z3)
	}

	// Increase threat for zone 2
	manager.TimedThreat(2, 1.0)
	if z2 := manager.GetZoneThreat(2); z2 != 1.0 {
		t.Errorf("Expected zone 2 threat 1.0, got %.2f", z2)
	}

	// Zone 1 should decay: previous 2.0 - 1.0/mapFactor = 1.0
	if z1 := manager.GetZoneThreat(1); z1 != 1.0 {
		t.Errorf("Expected zone 1 threat 1.0 after decay, got %.2f", z1)
	}
}

func TestThreatManager_AdvanceMap(t *testing.T) {
	manager := GetManager()
	manager.Reset()

	manager.RegisterZone(&Threat{ZoneID: 1})
	manager.RegisterZone(&Threat{ZoneID: 2})

	// Increase threat with mapFactor = 1
	manager.TimedThreat(1, 2.0)
	if z2 := manager.GetZoneThreat(2); z2 != 0 {
		t.Errorf("Expected zone 2 threat 0 after decay, got %.2f", z2)
	}

	// Advance map, mapFactor = 2
	manager.AdvanceMap()

	// Now decay should be halved
	manager.TimedThreat(1, 2.0) // z1 +=2, z2 -= 2/2 =1
	if z2 := manager.GetZoneThreat(2); z2 != 0 {
		t.Errorf("Expected zone 2 threat 0 after decay (clamped), got %.2f", z2)
	}
}

func TestThreatManager_Reset(t *testing.T) {
	manager := GetManager()
	manager.Reset()

	manager.RegisterZone(&Threat{ZoneID: 1})
	manager.TimedThreat(1, 5.0)

	manager.Reset()
	if len(manager.zones) != 0 {
		t.Errorf("Expected 0 zones after reset, got %d", len(manager.zones))
	}
	if manager.mapFactor != 1.0 {
		t.Errorf("Expected mapFactor 1.0 after reset, got %.2f", manager.mapFactor)
	}
}

func TestThreatManager_AutoCreateZone(t *testing.T) {
	manager := GetManager()
	manager.Reset()

	// TimedThreat on a non-registered zone should auto-create
	manager.TimedThreat(42, 3.0)
	if threat := manager.GetZoneThreat(42); threat != 3.0 {
		t.Errorf("Expected auto-created zone 42 threat 3.0, got %.2f", threat)
	}
}
