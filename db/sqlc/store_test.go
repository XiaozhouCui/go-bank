package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	// reuse testDB from main_test.go
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	// run n concurrent transfer transactions
	n := 5
	amount := int64(10)

	// create channels to connect concurrent goroutines
	// 1st channel to receive the error from each goroutine
	errs := make(chan error)
	// 2nd channel to receive the TransferTxResult from each goroutine
	results := make(chan TransferTxResult)

	// start new go routine for each concurrent transfer
	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err       // send error to errs channel
			results <- result // send result to results channel
		}()
	}

	// checkout results
	for i := 0; i < n; i++ {
		err := <-errs       // receive error from errs channel
		result := <-results // receive result from results channel

		// use testify to make assertions
		require.NoError(t, err)
		require.NotEmpty(t, result)

		// checkout transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)        // auto increment field
		require.NotZero(t, transfer.CreatedAt) // timestamp field
		// check if record is created in db
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check from entry
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount) // money is going out
		require.NotZero(t, fromEntry.ID)            // auto increment field
		require.NotZero(t, fromEntry.CreatedAt)     // timestamp field
		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		// check to entry
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount) // money is going in
		require.NotZero(t, toEntry.ID)           // auto increment field
		require.NotZero(t, toEntry.CreatedAt)    // timestamp field
		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// @TODO: check account balances
	}
}
