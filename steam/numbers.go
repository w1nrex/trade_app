package steam

func CheckNumbers(inv *InventoryResponse) []Skins {
	var skins []Skins
	seen := make(map[string]bool)
	index := 1

	for _, desc := range inv.Descriptions {
		if desc.Marketable == 1 && isWeaponSkin(desc.Type) && !seen[desc.ClassID] {
			skins = append(skins, Skins{
				Index:          index,
				Name:           desc.Name,
				MarketHashName: desc.MarketHashName,
				Type:           desc.Type,
			})
			seen[desc.ClassID] = true
			index++
		}
	}

	return skins
}
