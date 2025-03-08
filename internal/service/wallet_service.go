package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"github.com/sunriseex/test_wallet/internal/logger"
	"github.com/sunriseex/test_wallet/internal/model"
)

const (
	maxRetries         = 5
	serializationError = "40001"
	retryDelayBase     = 100 * time.Millisecond
)

type WalletService interface {
	GetBalance(ctx context.Context, walletID string) (model.Wallet, error)
	Deposit(ctx context.Context, walletID string, amount decimal.Decimal) error
	Withdraw(ctx context.Context, walletID string, amount decimal.Decimal) error
}

type WalletServiceImpl struct {
	db *sql.DB
}

func NewWalletService(db *sql.DB) *WalletServiceImpl {
	return &WalletServiceImpl{
		db: db,
	}
}

func (s *WalletServiceImpl) GetBalance(ctx context.Context, walletID string) (model.Wallet, error) {
	if _, err := uuid.Parse(walletID); err != nil {
		logger.Log.Errorf("Неверный формат UUID: %s", walletID)
		return model.Wallet{}, errors.New("invalid wallet ID format")

	}
	var wallet model.Wallet

	logger.Log.Info("Запрос к базе данных для получения баланса")
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

func (s *WalletServiceImpl) Deposit(ctx context.Context, walletID string, amount decimal.Decimal) error {

	if _, err := uuid.Parse(walletID); err != nil {
		logger.Log.Errorf("Неверный формат UUID: %s", walletID)
		return errors.New("invalid wallet ID format")
	}

	if amount.IsZero() {
		return nil
	}

	logger.Log.Infof("Попытка депозита: wallet_id=%s, amount=%s", walletID, amount)
	return s.updateBalance(ctx, walletID, amount)
}

func (s *WalletServiceImpl) Withdraw(ctx context.Context, walletID string, amount decimal.Decimal) error {
	if _, err := uuid.Parse(walletID); err != nil {
		logger.Log.Errorf("Неверный формат UUID: %s", walletID)
		return errors.New("invalid wallet ID format")
	}
	logger.Log.Infof("Попытка снятия: wallet_id=%s, amount=%s", walletID, amount)
	return s.updateBalance(ctx, walletID, amount.Neg())
}

func (s *WalletServiceImpl) updateBalance(ctx context.Context, walletID string, change decimal.Decimal) error {
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

func (s *WalletServiceImpl) executeWithRetry(ctx context.Context, fn func(*sql.Tx) error) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {

		if ctx.Err() != nil {
			return fmt.Errorf("operation canceled: %w", ctx.Err())
		}

		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
		})
		if err != nil {

			if isRetriableError(err) {
				lastErr = err
				logRetry(i, err)
				time.Sleep(calculateDelay(i))
				continue
			}
			return fmt.Errorf("non-retriable begin error: %w", err)
		}

		if err := fn(tx); err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logger.Log.Errorf("Rollback failed: %v", rbErr)
			}

			if !isRetriableError(err) {
				return fmt.Errorf("non-retriable error: %w", err)
			}

			lastErr = err
			logRetry(i, err)
			time.Sleep(calculateDelay(i))
			continue
		}

		if err := tx.Commit(); err != nil {
			if isRetriableError(err) {
				lastErr = err
				logRetry(i, err)
				time.Sleep(calculateDelay(i))
				continue
			}
			return fmt.Errorf("commit failed: %w", err)
		}

		return nil
	}

	return fmt.Errorf("max retries (%d) reached. Last error: %w", maxRetries, lastErr)
}

func createWallet(tx *sql.Tx, walletID string, balance decimal.Decimal) error {
	_, err := uuid.Parse(walletID)
	if err != nil {
		walletID = uuid.NewString()

	}
	logger.Log.Infof("Создание нового кошелька: wallet_id=%s, balance=%s", walletID, balance)

	queryInsert := `
    INSERT INTO wallet_db (wallet_id, balance)
    VALUES ($1, $2)`
	_, err = tx.Exec(queryInsert, walletID, balance)

	return err

}

func calculateDelay(attempt int) time.Duration {
	return time.Duration(attempt+1) * retryDelayBase
}

func logRetry(attempt int, err error) {
	logger.Log.Warnf("Retry attempt %d/%d. Reason: %v",
		attempt+1,
		maxRetries,
		err,
	)
}
func isNetError(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
func isRetriableError(err error) bool {

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {

		retriableCodes := map[string]struct{}{
			"40001": {},
			"40P01": {},
			"08006": {},
		}
		_, ok := retriableCodes[pgErr.Code]
		return ok
	}

	if isNetError(err) {
		return true
	}

	if err.Error() == "sql: transaction has already been committed or rolled back" {
		return true
	}

	return false
}
