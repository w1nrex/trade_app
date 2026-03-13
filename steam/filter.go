package steam

import (
	"strings"
)

func isWeaponSkin(itemType string) bool {
	excludes := []string{
		"Container", "Sticker", "Graffiti", "Music Kit",
		"Patch", "Tool", "Key", "Pass", "Gift", "Tag",
	}

	for _, ex := range excludes {
		if strings.Contains(itemType, ex) {
			return false
		}
	}

	weapons := []string{
		"Rifle", "Pistol", "SMG", "Sniper Rifle",
		"Shotgun", "Machinegun", "Knife",
	}

	for _, weapon := range weapons {
		if strings.Contains(itemType, weapon) {
			return true
		}
	}

	return false
}
