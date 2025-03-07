package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/sunriseex/test_wallet/internal/service"
)

type RequestBody struct {
	WalletID      string          `json:"walletId"`
	OperationType string          `json:"operationType"`
	Amount        decimal.Decimal `json:"amount"`
}
type WalletHandler struct {
	Logger        *logrus.Logger
	WalletService *service.WalletService
}

func NewWalletHandler(logger *logrus.Logger, walletService *service.WalletService) *WalletHandler {

	return &WalletHandler{Logger: logger,
		WalletService: walletService,
	}
}

func (h *WalletHandler) CreateOrUpdateWallet(w http.ResponseWriter, r *http.Request) {
	var req RequestBody

	if req.Amount.IsNegative() {
		http.Error(w, "Сумма не может быть отрицательной", http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.Logger.WithError(err).Error("Ошибка декодирования запроса: ", err)
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	switch req.OperationType {
	case "DEPOSIT":
		err = h.WalletService.Deposit(ctx, req.WalletID, req.Amount)
		if err != nil {
			h.Logger.WithError(err).Error("Ошибка при депозите: ", err)
			http.Error(w, "Ошибка депозита", http.StatusInternalServerError)
			return
		}
		h.Logger.Info("Депозит успешно выполнен")
	case "WITHDRAW":
		err = h.WalletService.Withdraw(ctx, req.WalletID, req.Amount)
		if err != nil {
			h.Logger.WithError(err).Error("Ошибка при снятии средств: ", err)
			http.Error(w, "Ошибка снятия средств", http.StatusInternalServerError)
			return
		}
		h.Logger.Info("Снятие средств успешно выполнено")
	default:
		h.Logger.Warnf("Неверный тип операции: %s", req.OperationType)
		http.Error(w, "Неверный тип операции", http.StatusBadRequest)
		return

	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Операция выполнена успешно"))

}

func (h *WalletHandler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletID := vars["walletId"]

	ctx := context.Background()
	wallet, err := h.WalletService.GetBalance(ctx, walletID)
	if err != nil {
		h.Logger.WithError(err).Error("Ошибка получения баланса: ", err)
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, "Кошелёк не найден", http.StatusNotFound)
			return
		}
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	h.Logger.Info("Баланс кошелька запрошен")

	resp := struct {
		WalletID string          `json:"walletId"`
		Balance  decimal.Decimal `json:"balance"`
	}{
		WalletID: walletID,
		Balance:  wallet.Balance,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}
