package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
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
	WalletService service.WalletService
}

func NewWalletHandler(
	logger *logrus.Logger,
	walletService service.WalletService,
) *WalletHandler {
	return &WalletHandler{
		Logger:        logger,
		WalletService: walletService,
	}
}

func (h *WalletHandler) CreateOrUpdateWallet(w http.ResponseWriter, r *http.Request) {
	var req RequestBody

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.WithError(err).Error("Ошибка декодирования запроса")
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(req.WalletID); err != nil {
		h.Logger.WithError(err).Errorf("Invalid wallet ID: %s", req.WalletID)
		http.Error(w, "Invalid wallet ID format", http.StatusBadRequest)
		return
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		h.Logger.Error("Сумма должна быть положительной")
		http.Error(w, "Сумма должна быть положительной", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	switch req.OperationType {
	case "DEPOSIT":
		if err := h.WalletService.Deposit(ctx, req.WalletID, req.Amount); err != nil {
			h.Logger.WithError(err).Errorf("Ошибка при депозите: WalletID=%s, Amount=%s", req.WalletID, req.Amount)
			http.Error(w, "Ошибка депозита", http.StatusBadRequest)
			return
		}
	case "WITHDRAW":
		if err := h.WalletService.Withdraw(ctx, req.WalletID, req.Amount); err != nil {
			h.Logger.WithError(err).Errorf("Ошибка при снятии средств: WalletID=%s, Amount=%s", req.WalletID, req.Amount)
			http.Error(w, "Ошибка снятия средств", http.StatusBadRequest)
			return
		}
		h.Logger.Infof("Снятие средств успешно выполнено: WalletID=%s, Amount=%s", req.WalletID, req.Amount)
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

	if _, err := uuid.Parse(walletID); err != nil {
		h.Logger.WithError(err).Errorf("Invalid wallet ID: %s", walletID)
		http.Error(w, "Invalid wallet ID format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	wallet, err := h.WalletService.GetBalance(ctx, walletID)
	if err != nil {
		h.Logger.WithError(err).Errorf("GetBalance error: %s", walletID)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Wallet not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	resp := struct {
		WalletID string          `json:"walletId"`
		Balance  decimal.Decimal `json:"balance"`
	}{
		WalletID: wallet.WalletID,
		Balance:  wallet.Balance,
	}
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Logger.WithError(err).Error("Ошибка кодирования ответа")
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return

	}

}
