package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sunriseex/test_wallet/internal/config"
)

func main() {
	cfg := config.LoadConfig()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Приложение test_wallet запущено!")
	})

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	fmt.Printf("Запуск сервера на порту: %s", cfg.AppPort)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
