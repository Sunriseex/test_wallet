package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Wallet struct {
	WalletID  string          `json:"walletId"`
	Balance   decimal.Decimal `json:"balance"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}
