package db

import (
	"context"
	"testing"
	"time"

	"github.com/XiaozhouCui/go-bank/db/util"
	"github.com/stretchr/testify/require"
)

// TestCreateTransfer creates a transfer and verifies the returned transfer
func TestCreateTransfer(t *testing.T) {
	// Create two random accounts
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	// Create a random transfer
	arg := CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        util.RandomMoney(),
	}

	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)
}

// TestGetTransfer gets a transfer and verifies the returned transfer
func TestGetTransfer(t *testing.T) {
	// Create two random accounts
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	// Create a random transfer
	arg := CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        util.RandomMoney(),
	}

	transfer1, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer1)

	// Get the created transfer
	transfer2, err := testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.Equal(t, transfer1.Amount, transfer2.Amount)
	require.WithinDuration(t, transfer1.CreatedAt, transfer2.CreatedAt, time.Second)
}

// TestListTransfers lists transfers and verifies the returned transfers
func TestListTransfers(t *testing.T) {
	// Create two random accounts
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	// Create 10 random transfers between the two accounts
	arg := CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        util.RandomMoney(),
	}

	for i := 0; i < 10; i++ {
		_, err := testQueries.CreateTransfer(context.Background(), arg)
		require.NoError(t, err)
	}

	// Skip the first 5 transfers and get the next 5 transfers
	arg2 := ListTransfersParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Limit:         5,
		Offset:        5,
	}

	transfers, err := testQueries.ListTransfers(context.Background(), arg2)
	require.NoError(t, err)
	require.Len(t, transfers, 5)

	for _, transfer := range transfers {
		require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
		require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
		require.Equal(t, arg.Amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)
	}
}
