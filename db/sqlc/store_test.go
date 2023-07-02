package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	// reuse testDB from main_test.go
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

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
	existed := make(map[int]bool)
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

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		// check accounts' balance
		fmt.Println(">> tx:", fromAccount.Balance, toAccount.Balance)
		diff1 := account1.Balance - fromAccount.Balance // money going out of account1
		diff2 := toAccount.Balance - account2.Balance   // money going into account2
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)         // positive number
		require.True(t, diff1%amount == 0) // 1 * amount, 2 * amount, 3 * amount, ..., n * amount

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)  // k is unique for each transaction: 1, 2, 3, ..., n
		require.NotContains(t, existed, k) // existed does not contain k
		existed[k] = true                  // then add k to the map at the end of for loop
	}

	// check the final updated balance of accounts
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {
	// reuse testDB from main_test.go
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	// run 10 concurrent transfer transactions: 5 from acc1 to acc2, 5 from acc2 to acc1
	n := 10
	amount := int64(10)
	errs := make(chan error)

	// start new go routine for each concurrent transfer
	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID
		// for odd numbers: reverse the transfer
		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}
		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err // send error to errs channel
		}()
	}

	// checkout results
	for i := 0; i < n; i++ {
		err := <-errs // receive error from errs channel
		require.NoError(t, err)
	}

	// check the final updated balance of accounts
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)
}
