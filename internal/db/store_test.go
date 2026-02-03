package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	// Generate unique email based on time to avoid collision on repeated local runs
	uniqueSuffix := time.Now().UnixNano()

	user1, err := testStore.CreateUser(context.Background(), CreateUserParams{
		Username:       fmt.Sprintf("user1_%d", uniqueSuffix),
		Email:          fmt.Sprintf("user1_%d@example.com", uniqueSuffix),
		HashedPassword: "secret",
	})
	require.NoError(t, err)

	user2, err := testStore.CreateUser(context.Background(), CreateUserParams{
		Username:       fmt.Sprintf("user2_%d", uniqueSuffix),
		Email:          fmt.Sprintf("user2_%d@example.com", uniqueSuffix),
		HashedPassword: "secret",
	})
	require.NoError(t, err)

	account1, err := testStore.CreateAccount(context.Background(), CreateAccountParams{
		UserID:   user1.ID,
		Balance:  100,
		Currency: "USD",
	})
	require.NoError(t, err)

	account2, err := testStore.CreateAccount(context.Background(), CreateAccountParams{
		UserID:   user2.ID,
		Balance:  100,
		Currency: "USD",
	})
	require.NoError(t, err)

	// CONCURRENCY TEST: Execute n concurrent transfer transactions
	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := testStore.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	// Check results
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = testStore.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		// check balances
		fmt.Println(">> tx:", fromAccount.Balance, toAccount.Balance)

		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// Check final updated balances
	updatedAccount1, err := testStore.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testStore.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance)
}
