package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lwandokasuba/golang-sqlc/internal/db"
	"github.com/stretchr/testify/require"
)

func TestGetUserWithAccounts(t *testing.T) {
	svc := NewService(testStore)
	uniqueSuffix := time.Now().UnixNano()

	// 1. Create User via Store (to look more like setup)
	userParams := db.CreateUserParams{
		Username:       fmt.Sprintf("user_eager_%d", uniqueSuffix),
		Email:          fmt.Sprintf("user_eager_%d@example.com", uniqueSuffix),
		HashedPassword: "secret",
	}
	user, err := testStore.CreateUser(context.Background(), userParams)
	require.NoError(t, err)

	// 2. Create Accounts
	acc1Params := db.CreateAccountParams{
		UserID:   user.ID,
		Balance:  500,
		Currency: "USD",
	}
	_, err = testStore.CreateAccount(context.Background(), acc1Params)
	require.NoError(t, err)

	acc2Params := db.CreateAccountParams{
		UserID:   user.ID,
		Balance:  1000,
		Currency: "EUR",
	}
	_, err = testStore.CreateAccount(context.Background(), acc2Params)
	require.NoError(t, err)

	// 3. Test GetUser WITHOUT embed
	res, err := svc.GetUser(context.Background(), user.ID, GetUserOptions{IncludeAccounts: false})
	require.NoError(t, err)
	require.Equal(t, user.ID, res.ID)
	require.Empty(t, res.Accounts)

	// 4. Test GetUser WITH embed
	resEmbed, err := svc.GetUser(context.Background(), user.ID, GetUserOptions{IncludeAccounts: true})
	require.NoError(t, err)
	require.Equal(t, user.ID, resEmbed.ID)
	require.Len(t, resEmbed.Accounts, 2)

	// Validate account contents
	var foundUSD, foundEUR bool
	for _, acc := range resEmbed.Accounts {
		if acc.Currency == "USD" {
			foundUSD = true
			require.Equal(t, int64(500), acc.Balance)
		}
		if acc.Currency == "EUR" {
			foundEUR = true
			require.Equal(t, int64(1000), acc.Balance)
		}
	}
	require.True(t, foundUSD)
	require.True(t, foundEUR)
}
