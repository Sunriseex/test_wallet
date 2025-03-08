package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sunriseex/test_wallet/internal/config"
	"github.com/sunriseex/test_wallet/internal/db"
	"github.com/sunriseex/test_wallet/internal/handler"
	"github.com/sunriseex/test_wallet/internal/logger"
	"github.com/sunriseex/test_wallet/internal/middleware"
	"github.com/sunriseex/test_wallet/internal/service"
)

func main() {

	cfg := config.LoadConfig()

	logger.InitLogger()

	logger.Log.Info("Сервер запускается...")

	database := db.InitDB(cfg)
	db.InitSchema(database)

	walletService := service.NewWalletService(database)
	workerPool := service.NewWorkerPool(walletService, 50, 1000)

	r := mux.NewRouter()

	walletHandler := handler.NewWalletHandler(logger.Log, walletService)

	r.Use(middleware.LoggerMiddleware)
	r.HandleFunc("/api/v1/wallet", walletHandler.CreateOrUpdateWallet).Methods("POST")
	r.HandleFunc("/api/v1/wallets/{walletId}", walletHandler.GetWalletBalance).Methods("GET")

	addr := fmt.Sprintf(":%s", cfg.AppPort)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	logger.Log.Infof("Сервер запущен на порту: %s", cfg.AppPort)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	workerPool.Shutdown()

	database.Close()
	logger.Log.Info("База данных закрыта успешно")

	logger.Log.Info("Логгер остановлен успешно")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatalf("Ошибка при остановке сервера: %v", err)
	}

	logger.Log.Info("Сервер остановлен успешно")

}
