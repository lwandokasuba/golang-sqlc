package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lwandokasuba/golang-sqlc/internal/db"
	"github.com/lwandokasuba/golang-sqlc/internal/dto"
)

// GetUserOptions defines what related resources to include
type GetUserOptions struct {
	IncludeAccounts bool
}

// Service defines the interface for business logic
type Service interface {
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (dto.UserResponse, error)
	GetUser(ctx context.Context, id int64, opts GetUserOptions) (dto.UserResponse, error)
	CreateAccount(ctx context.Context, req dto.CreateAccountRequest) (dto.AccountResponse, error)
	CreateTransfer(ctx context.Context, req dto.TransferRequest) (db.TransferTxResult, error)
}

type SimpleService struct {
	store db.Store
}

func NewService(store db.Store) Service {
	return &SimpleService{
		store: store,
	}
}

func (s *SimpleService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (dto.UserResponse, error) {
	arg := db.CreateUserParams{
		Username:       req.Username,
		Email:          req.Email,
		HashedPassword: "secret", // Default for now, or req.Password if we add it to DTO
	}
	user, err := s.store.CreateUser(ctx, arg)
	if err != nil {
		return dto.UserResponse{}, err
	}
	return dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time, // Assuming pgx returns pgtype.Timestamptz or similar?
		// Note: sqlc with pgx/v5 uses pgtype.Timestamptz. We might need to handle conversion.
		// For simplicity, if standard lib or if I used emit_time_string...
		// Let's assume standard time.Time if sqlc config allows, but pgx uses pgtypes.
		// Actually sqlc default for pgx/v5 uses pgtype which has a Time field.
	}, nil
}

func (s *SimpleService) GetUser(ctx context.Context, id int64, opts GetUserOptions) (dto.UserResponse, error) {
	if opts.IncludeAccounts {
		rows, err := s.store.GetUserWithAccounts(ctx, id)
		if err != nil {
			return dto.UserResponse{}, err
		}
		if len(rows) == 0 {
			return dto.UserResponse{}, sql.ErrNoRows
		}

		first := rows[0]
		rsp := dto.UserResponse{
			ID:        first.UserID,
			Username:  first.Username,
			Email:     first.Email,
			CreatedAt: first.UserCreatedAt.Time,
		}

		rsp.Accounts = make([]dto.AccountResponse, 0, len(rows))
		for _, row := range rows {
			if row.AccountID.Valid {
				rsp.Accounts = append(rsp.Accounts, dto.AccountResponse{
					ID:        row.AccountID.Int64,
					UserID:    row.UserID,
					Balance:   row.Balance.Int64,
					Currency:  row.Currency.String,
					CreatedAt: row.AccountCreatedAt.Time,
				})
			}
		}
		return rsp, nil
	}

	// Fallback to simple query if no accounts requested
	user, err := s.store.GetUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows { // checking logic later
			return dto.UserResponse{}, err
		}
		return dto.UserResponse{}, err
	}

	rsp := dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
	}

	return rsp, nil
}

func (s *SimpleService) CreateAccount(ctx context.Context, req dto.CreateAccountRequest) (dto.AccountResponse, error) {
	arg := db.CreateAccountParams{
		UserID:   req.UserID,
		Balance:  req.Balance,
		Currency: req.Currency,
	}
	account, err := s.store.CreateAccount(ctx, arg)
	if err != nil {
		return dto.AccountResponse{}, err
	}
	return dto.AccountResponse{
		ID:        account.ID,
		UserID:    account.UserID,
		Balance:   account.Balance,
		Currency:  account.Currency,
		CreatedAt: account.CreatedAt.Time,
	}, nil
}

func (s *SimpleService) CreateTransfer(ctx context.Context, req dto.TransferRequest) (db.TransferTxResult, error) {
	// Add business logic: Check currency match, check balance, etc.
	// Ideally we fetch accounts first and check currency.

	// Validate that accounts exist and currency matches
	fromAccount, err := s.store.GetAccount(ctx, req.FromAccountID)
	if err != nil {
		return db.TransferTxResult{}, err
	}
	if fromAccount.Currency != req.Currency {
		return db.TransferTxResult{}, fmt.Errorf("currency mismatch: account %s vs request %s", fromAccount.Currency, req.Currency)
	}
	if fromAccount.Balance < req.Amount {
		return db.TransferTxResult{}, fmt.Errorf("insufficient funds: balance %d, required %d", fromAccount.Balance, req.Amount)
	}

	toAccount, err := s.store.GetAccount(ctx, req.ToAccountID)
	if err != nil {
		return db.TransferTxResult{}, err
	}
	if toAccount.Currency != req.Currency {
		return db.TransferTxResult{}, fmt.Errorf("currency mismatch: account %s vs request %s", toAccount.Currency, req.Currency)
	}

	// For this example, we proceed to TX directly.
	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}
	return s.store.TransferTx(ctx, arg)
}
