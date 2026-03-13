package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"tradeapp/server"
	"tradeapp/steam"
)

const SteamID = "76561199159548432"

func main() {
	srv := server.Server()
	go func() {
		fmt.Println("Сервер запущен на :8080")
		if err := http.ListenAndServe(":8080", srv.Handler); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(500 * time.Millisecond)
	steam.BaseURL = "http://localhost:8080"
	if err := steam.MainSteam(SteamID, func() {
		fmt.Println("🔄 Повторный анализ рынка...")
	}); err != nil {
		log.Fatalf("Ошибка: %v", err)
	}
}
