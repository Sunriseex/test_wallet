package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/sunriseex/test_wallet/internal/model"
)

type RequestBody struct {
	WalletID      string          `json:"walletId"`
	OperationType string          `json:"operationType"`
	Amount        decimal.Decimal `json:"amount"`
}
type WalletHandler struct {
	Logger *logrus.Logger
}

func NewWalletHandler(logger *logrus.Logger) *WalletHandler {

	return &WalletHandler{Logger: logger}
}

func (h *WalletHandler) CreateOrUpdateWallet(w http.ResponseWriter, r *http.Request) {
	var req RequestBody
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		h.Logger.Error("Error decoding request body:", err)
		http.Error(w, "Bad request format", http.StatusBadRequest)
		return
	}

	h.Logger.Infof("Получена операция %s для кошелька %s на сумму %s", req.OperationType, req.WalletID, req.Amount.String())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Операция выполнена"))

}

func (h *WalletHandler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletID := vars["walletId"]
	wallet := model.Wallet{
		WalletID:  walletID,
		Balance:   decimal.NewFromInt(0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)

}
