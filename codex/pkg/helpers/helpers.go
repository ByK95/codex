package helpers

import (
	"codex/pkg/crafting"
	"codex/pkg/equipment"
	"codex/pkg/iterator"
	"codex/pkg/store"
	"fmt"
	"math/rand"
	"time"
)

var upgrades string = "upgrades"
var upgradesIter *iterator.Iterator[string]

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetUpgrades() []string {
	crafter, ok := crafting.Get(upgrades)
	if !ok {
		return []string{}
	}
	equipments := equipment.GetManager().GetAllEquippedItems()

	resultSet := make(map[string]struct{})
	equippedSet := make(map[string]struct{}, len(equipments))

	for _, itemID := range equipments {
		equippedSet[itemID] = struct{}{}
		craftables := crafter.FindByRequirement(itemID)
		for _, c := range craftables {

			if _, already := equippedSet[c.ID]; already {
                continue
            }

			resultSet[c.ID] = struct{}{}
		}
	}
	

	if equipment.GetManager().HasAnyEmptySlot() {
		craftables := crafter.FindByRequirement("")
		for _, c := range craftables {
			
			slotType := getEquipmentSlotType(c.ID)
			if equipment.GetManager().GetSlotAvailability(slotType) == 0{
				continue
			}

			if _, already := equippedSet[c.ID]; already {
                continue
            }

			resultSet[c.ID] = struct{}{}
		}
	}

	results := make([]string, 0, len(resultSet))
	for id := range resultSet {
		results = append(results, id)
	}

	return results
}

func getEquipmentSlotType(itemID string) string {
	keyName := fmt.Sprintf("%s.slot_type", itemID)
	return store.GetStore().GetString(keyName)
}

func UpgrageItem(itemID string) bool {
	crafter, _ := crafting.Get(upgrades)
	craftable, ok := crafter.GetCraftable(itemID)
	slotType := getEquipmentSlotType(itemID)
	if !ok {
		return false
	}

	req := craftable.Requirements[0].ID 
	if req == "" && equipment.GetManager().GetSlotAvailability(slotType) !=0 {
		equipment.GetManager().EquipItem(slotType, itemID)
		return true
	}
	ok = equipment.GetManager().UnequipItem(slotType, req)
	if !ok {
		return false
	}
	ok = equipment.GetManager().EquipItem(slotType, itemID)
	return ok
}

func getUpgradeSelections(count int) []string {
	upgrades := GetUpgrades()
	if len(upgrades) == 0 || count <= 0 {
		return nil
	}

	cp := append([]string(nil), upgrades...)
	rand.Shuffle(len(cp), func(i, j int) { cp[i], cp[j] = cp[j], cp[i] })

	if count > len(cp) {
		count = len(cp)
	}
	return cp[:count]
}

func GetUpgradeSelections(count int){
	selections := getUpgradeSelections(count)
	upgradesIter = iterator.NewIterator(selections)
}

func GetNextSelections() string{
	if upgradesIter == nil {
		return ""
	}
	val, _ := upgradesIter.Next()
	return val
}