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

// NewStore creates a new store.
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
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
	// Execute the callback function (arg) with the query object.
	err = fn(q)
	// If there is an error, rollback the transaction.
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
