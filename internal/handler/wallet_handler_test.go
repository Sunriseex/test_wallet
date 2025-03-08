package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"

	"github.com/sunriseex/test_wallet/internal/model"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidOperation  = errors.New("invalid operation")
)

type mockWalletService struct {
	mock.Mock
}

func (m *mockWalletService) Deposit(ctx context.Context, walletID string, amount decimal.Decimal) error {
	return nil
}

func (m *mockWalletService) Withdraw(ctx context.Context, walletID string, amount decimal.Decimal) error {
	balance := decimal.NewFromInt(500)
	if amount.GreaterThan(balance) {
		return ErrInsufficientFunds
	}
	return nil
}

func (m *mockWalletService) GetBalance(ctx context.Context, walletID string) (model.Wallet, error) {
	return model.Wallet{
		WalletID:  walletID,
		Balance:   decimal.NewFromInt(500),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func TestCreateOrUpdateWallet(t *testing.T) {

	svc := &mockWalletService{}
	logger := logrus.New()
	handler := NewWalletHandler(logger, svc)

	testCases := []struct {
		name           string
		requestBody    string
		expectedStatus int
		mockResponse   error
	}{
		{
			name: "Пополнение баланса",
			requestBody: `{
				"walletId": "550e8400-e29b-41d4-a716-446655440000",
				"operationType": "DEPOSIT",
				"amount": "100.50"
			}`,
			expectedStatus: http.StatusOK,
			mockResponse:   nil,
		},
		{
			name: "Снятие баланса (успешное)",
			requestBody: `{
				"walletId": "550e8400-e29b-41d4-a716-446655440000",
				"operationType": "WITHDRAW",
				"amount": "50.00"
			}`,
			expectedStatus: http.StatusOK,
			mockResponse:   nil,
		},
		{
			name: "Снятие больше баланса",
			requestBody: `{
				"walletId": "550e8400-e29b-41d4-a716-446655440000",
				"operationType": "WITHDRAW",
				"amount": "1000000"
			}`,
			expectedStatus: http.StatusBadRequest,
			mockResponse:   ErrInsufficientFunds,
		},
		{
			name: "Некорректная операция",
			requestBody: `{
				"walletId": "550e8400-e29b-41d4-a716-446655440000",
				"operationType": "UNKNOWN",
				"amount": "50.00"
			}`,
			expectedStatus: http.StatusBadRequest,
			mockResponse:   ErrInvalidOperation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/wallet", strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			svc.On("ProcessTransaction", mock.Anything, mock.Anything, mock.Anything).Return(tc.mockResponse)

			handler.CreateOrUpdateWallet(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("%s: Expected %d, got %d", tc.name, tc.expectedStatus, w.Code)
			}
		})
	}

}

func TestGetWalletBalance(t *testing.T) {
	svc := &mockWalletService{}
	logger := logrus.New()
	handler := NewWalletHandler(logger, svc)

	req := httptest.NewRequest("GET", "/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000", nil)
	req = mux.SetURLVars(req, map[string]string{"walletId": "550e8400-e29b-41d4-a716-446655440000"})
	w := httptest.NewRecorder()
	handler.GetWalletBalance(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}
func TestInvalidWalletID(t *testing.T) {
	svc := &mockWalletService{}
	logger := logrus.New()
	handler := NewWalletHandler(logger, svc)
	req := httptest.NewRequest("GET", "/api/v1/wallets/invalid_id", nil)
	req = mux.SetURLVars(req, map[string]string{"walletId": "invalid_id"})
	w := httptest.NewRecorder()
	handler.GetWalletBalance(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestWalletNotFound(t *testing.T) {
	svc := &mockWalletService{}
	logger := logrus.New()
	handler := NewWalletHandler(logger, svc)

	req := httptest.NewRequest("GET", "/api/v1/wallets/550e8400-e29b-41d4-446655440000", nil)
	req = mux.SetURLVars(req, map[string]string{"walletId": "550e8400-e29b-41d4-446655440000"})
	w := httptest.NewRecorder()
	handler.GetWalletBalance(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}
