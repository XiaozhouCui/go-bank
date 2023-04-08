package db

import "database/sql"

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
