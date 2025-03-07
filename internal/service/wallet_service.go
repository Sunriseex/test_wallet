package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"github.com/sunriseex/test_wallet/internal/logger"
	"github.com/sunriseex/test_wallet/internal/model"
)

const (
	maxRetries         = 3
	serializationError = "40001"
	retryDelay         = 100 * time.Millisecond
)

type WalletService struct {
	db *sql.DB
}

func NewWalletService(db *sql.DB) *WalletService {
	return &WalletService{
		db: db,
	}
}

func (s *WalletService) GetBalance(ctx context.Context, walletID string) (model.Wallet, error) {
	if _, err := uuid.Parse(walletID); err != nil {
		logger.Log.Errorf("Неверный формат UUID: %s", walletID)
		return model.Wallet{}, errors.New("invalid wallet ID format")
	}

	var wallet model.Wallet
	query := `
        SELECT wallet_id, balance, created_at, updated_at
        FROM wallet_db
        WHERE wallet_id = $1
    `
	row := s.db.QueryRowContext(ctx, query, walletID)
	err := row.Scan(&wallet.WalletID, &wallet.Balance, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Wallet{}, sql.ErrNoRows
		}
		return model.Wallet{}, err
	}
	return wallet, nil
}

func (s *WalletService) Deposit(ctx context.Context, walletID string, amount decimal.Decimal) error {
	if _, err := uuid.Parse(walletID); err != nil {
		logger.Log.Errorf("Неверный формат UUID: %s", walletID)
		return errors.New("invalid wallet ID format")
	}
	logger.Log.Infof("Попытка депозита: wallet_id=%s, amount=%s", walletID, amount)
	return s.updateBalance(ctx, walletID, amount)
}

func (s *WalletService) Withdraw(ctx context.Context, walletID string, amount decimal.Decimal) error {
	if _, err := uuid.Parse(walletID); err != nil {
		logger.Log.Errorf("Неверный формат UUID: %s", walletID)
		return errors.New("invalid wallet ID format")
	}
	logger.Log.Infof("Попытка снятия: wallet_id=%s, amount=%s", walletID, amount)
	return s.updateBalance(ctx, walletID, amount.Neg())
}

func (s *WalletService) updateBalance(ctx context.Context, walletID string, change decimal.Decimal) error {
	return s.executeWithRetry(ctx, func(tx *sql.Tx) error {
		var currentBalance decimal.Decimal
		var createdAt, updatedAt time.Time

		querySelect := `
            SELECT balance, created_at, updated_at
            FROM wallet_db
            WHERE wallet_id = $1
            FOR UPDATE`

		row := tx.QueryRowContext(ctx, querySelect, walletID)
		err := row.Scan(&currentBalance, &createdAt, &updatedAt)

		if errors.Is(err, sql.ErrNoRows) {
			newBalance := change
			if newBalance.IsNegative() {
				return errors.New("insufficient funds: wallet not found and negative deposit is not possible")
			}
			return createWallet(tx, walletID, newBalance)
		}
		if err != nil {
			return err
		}

		newBalance := currentBalance.Add(change)
		if newBalance.IsNegative() {
			return errors.New("insufficient funds")
		}

		queryUpdate := `
            UPDATE wallet_db
            SET balance = $1, updated_at = NOW()
            WHERE wallet_id = $2`

		_, err = tx.ExecContext(ctx, queryUpdate, newBalance, walletID)
		return err
	})
}

func (s *WalletService) executeWithRetry(ctx context.Context, fn func(*sql.Tx) error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		tx, beginErr := s.db.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelSerializable,
		})
		if beginErr != nil {
			return beginErr
		}

		err = fn(tx)
		if err == nil {
			return tx.Commit()
		}

		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Log.Errorf("Ошибка отката транзакции: %v", rbErr)
		}

		if !isRetriableError(err) {
			return err
		}

		logger.Log.Warnf("Повторная попытка (%d/%d). Ошибка: %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}
	return fmt.Errorf("достигнут максимум попыток: %w", err)
}
func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == serializationError {
		return true
	}
	return false
}

func createWallet(tx *sql.Tx, walletID string, balance decimal.Decimal) error {
	_, err := uuid.Parse(walletID)
	if err != nil {
		walletID = uuid.NewString()
	}

	queryInsert := `
    INSERT INTO wallet_db (wallet_id, balance)
    VALUES ($1, $2)`
	_, err = tx.Exec(queryInsert, walletID, balance)
	return err
}
