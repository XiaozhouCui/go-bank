package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db queries and transactions.
type Store struct {
	// composit: embed Queries struct to extend Store with all query methods.
	*Queries
	db *sql.DB
}

// NewStore creates a new store object.
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db), // New() is generated by sqlc
	}
}

// execTx executes a function within a database transaction.
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	// BeginTx starts a transaction, and returns a transaction object
	// The provided context is used until the transaction is committed or rolled back.
	tx, err := store.db.BeginTx(ctx, nil) // nil means use default tx options.
	if err != nil {
		return err
	}
	// Create a new query object with the transaction object.
	q := New(tx)
	// Execute the callback function (arg fn) with the query object.
	err = fn(q)
	// Check if there is a transaction error.
	if err != nil {
		// Rollback returns an error if the transaction has already been committed or rolled back.
		if rbErr := tx.Rollback(); rbErr != nil {
			// If there is an error rolling back, return both errors.
			return fmt.Errorf("tx error: %v, rb error: %v", err, rbErr)
		}
		// If there is no error rolling back, return the original error.
		return err
	}
	// If there is no error, commit the transaction.
	return tx.Commit()
}

// TransferTxParams contains the input parameters of the transfer transaction.
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult contains the result of the transfer transaction.
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`     // new transfer record
	FromAccount Account  `json:"from_account"` // amount after transfer
	ToAccount   Account  `json:"to_account"`   // amount after transfer
	FromEntry   Entry    `json:"from_entry"`   // new entry for from account, records the money is moving out
	ToEntry     Entry    `json:"to_entry"`     // new entry for to account, records the money is moving in
}

// TransferTx performs a transfer between two accounts within a database transaction.
// It creates a transfer record, add account entries, and update account balances within a single transaction.
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	// create an empty result
	var result TransferTxResult

	// create and run new database transaction
	err := store.execTx(ctx, func(q *Queries) error {
		// start the callback function "fn"
		var err error
		// create transfer record, using the generated query method "CreateTransfer" from sqlc
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		// add account entry for the FromAccount
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount, // negative value: money is moving out
		})
		if err != nil {
			return err
		}

		// add account entry for the ToAccount
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount, // positive value: money is moving in
		})
		if err != nil {
			return err
		}

		// update account balances
		account1, err := q.GetAccountForUpdate(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}

		result.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      arg.FromAccountID,
			Balance: account1.Balance - arg.Amount,
		})
		if err != nil {
			return err
		}

		account2, err := q.GetAccountForUpdate(ctx, arg.ToAccountID)
		if err != nil {
			return err
		}

		result.ToAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      arg.ToAccountID,
			Balance: account2.Balance + arg.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
