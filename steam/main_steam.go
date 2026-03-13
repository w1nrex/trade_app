package steam

import (
	"fmt"
	"time"
	"tradeapp/db"
)

func MainSteam(steamID string, f func()) error {
	fmt.Println("=== ТРЕКЕР ЦЕН CS2 ===")
	fmt.Println("Загружаю инвентарь...")
	inv, err := GetInventory(steamID)
	if err != nil {
		return fmt.Errorf("ошибка загрузки инвентаря: %v", err)
	}
	fmt.Printf("✅ Инвентарь загружен (%d предметов)\n", inv.TotalCount)

	allSkins := CheckNumbers(inv)
	if len(allSkins) == 0 {
		return fmt.Errorf("скины не найдены")
	}
	analysis, err := Analysis(allSkins)
	if err != nil {
		return fmt.Errorf("ошибка анализа рынка: %v", err)
	}
	db.InsertDB(analysis)

	go func() {
		for {
			f()
			time.Sleep(time.Minute * 5)
		}
	}()
	return err
}
