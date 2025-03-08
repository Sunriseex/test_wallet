package service

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Test Suite
type WalletServiceSuite struct {
	suite.Suite
	db      *sql.DB
	mock    sqlmock.Sqlmock
	service WalletService
}

func (s *WalletServiceSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	require.NoError(s.T(), err)
	s.db = db
	s.mock = mock
	s.service = NewWalletService(db)
}

func (s *WalletServiceSuite) TearDownTest() {
	s.db.Close()
}

func TestWalletServiceSuite(t *testing.T) {
	suite.Run(t, new(WalletServiceSuite))
}

func (s *WalletServiceSuite) TestDeposit_NewWallet() {
	walletID := "550e8400-e29b-41d4-a716-446655440000"
	amount := decimal.NewFromFloat(100.50)

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT balance, created_at, updated_at FROM wallet_db WHERE wallet_id = $1 FOR UPDATE`)).
		WithArgs(walletID).
		WillReturnError(sql.ErrNoRows)
	s.mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO wallet_db (wallet_id, balance) VALUES ($1, $2)`)).
		WithArgs(walletID, amount).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	err := s.service.Deposit(context.Background(), walletID, amount)

	assert.NoError(s.T(), err)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *WalletServiceSuite) TestDeposit_InvalidUUID() {
	err := s.service.Deposit(context.Background(), "invalid-uuid", decimal.NewFromInt(100))

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid wallet ID format")
}

func (s *WalletServiceSuite) TestWithdraw_InsufficientFunds() {
	walletID := "550e8400-e29b-41d4-a716-446655440000"
	initialBalance := decimal.NewFromInt(50)
	withdrawAmount := decimal.NewFromInt(100)

	s.mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{"balance", "created_at", "updated_at"}).
		AddRow(initialBalance, time.Now(), time.Now())
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT balance, created_at, updated_at FROM wallet_db WHERE wallet_id = $1 FOR UPDATE`)).
		WithArgs(walletID).
		WillReturnRows(rows)
	s.mock.ExpectRollback()

	err := s.service.Withdraw(context.Background(), walletID, withdrawAmount)

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "insufficient funds")
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *WalletServiceSuite) TestDeposit_RetryOnSerializationError() {
	walletID := "550e8400-e29b-41d4-a716-446655440000"
	amount := decimal.NewFromInt(100)

	s.mock.ExpectBegin().WillReturnError(&pgconn.PgError{Code: "40001"})
	s.mock.ExpectBegin().WillReturnError(&pgconn.PgError{Code: "40001"})
	s.mock.ExpectBegin().WillReturnError(&pgconn.PgError{Code: "40001"})

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT balance, created_at, updated_at FROM wallet_db WHERE wallet_id = $1 FOR UPDATE`)).
		WithArgs(walletID).
		WillReturnError(sql.ErrNoRows)
	s.mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO wallet_db (wallet_id, balance) VALUES ($1, $2)`)).
		WithArgs(walletID, amount).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	err := s.service.Deposit(context.Background(), walletID, amount)

	assert.NoError(s.T(), err)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *WalletServiceSuite) TestDeposit_ContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.service.Deposit(ctx, "550e8400-e29b-41d4-a716-446655440000", decimal.NewFromInt(100))

	assert.Error(s.T(), err)
	assert.True(s.T(), errors.Is(err, context.Canceled))
}

func (s *WalletServiceSuite) TestGetBalance_NotFound() {
	walletID := "550e8400-e29b-41d4-a716-446655440000"

	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT wallet_id, balance, created_at, updated_at FROM wallet_db WHERE wallet_id = $1`)).
		WithArgs(walletID).
		WillReturnError(sql.ErrNoRows)

	_, err := s.service.GetBalance(context.Background(), walletID)

	assert.Error(s.T(), err)
	assert.True(s.T(), errors.Is(err, sql.ErrNoRows))
}

func (s *WalletServiceSuite) TestDeposit_ZeroAmount() {
	err := s.service.Deposit(context.Background(), "550e8400-e29b-41d4-a716-446655440000", decimal.Zero)

	assert.NoError(s.T(), err)
}
