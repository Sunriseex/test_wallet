package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/sunriseex/test_wallet/internal/config"
	"github.com/sunriseex/test_wallet/internal/db"
	"github.com/sunriseex/test_wallet/internal/handler"
)

func main() {
	cfg := config.LoadConfig()

	logger := logrus.New()

	database := db.InitDB(cfg)
	db.InitSchema(database)

	r := mux.NewRouter()

	walletHandler := handler.NewWalletHandler(logger)

	r.HandleFunc("/api/v1/wallet", walletHandler.CreateOrUpdateWallet).Methods("POST")
	r.HandleFunc("/api/v1/wallets/{walletId}", walletHandler.GetWalletBalance).Methods("GET")

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Обработка запроса на корневой маршрут")
		fmt.Fprintf(w, "Приложение test_wallet запущено и работает!")
	}).Methods("GET")

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	logger.Infof("Запуск сервера на порту %s", cfg.AppPort)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
