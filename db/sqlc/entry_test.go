package db

import (
	"context"
	"testing"
	"time"

	"github.com/XiaozhouCui/go-bank/db/util"
	"github.com/stretchr/testify/require"
)

// createRandomEntry creates a random entry
func createRandomEntry(t *testing.T, account Account) Entry {
	arg := CreateEntryParams{
		AccountID: account.ID,
		Amount:    util.RandomMoney(),
	}

	entry, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry
}

// TestCreateEntry creates an entry and verifies the returned entry
func TestCreateEntry(t *testing.T) {
	account := createRandomAccount(t)
	createRandomEntry(t, account)
}

// TestGetEntry gets an entry and verifies the returned entry
func TestGetEntry(t *testing.T) {
	// Create a random account
	account := createRandomAccount(t)

	// Create a random entry
	entry1 := createRandomEntry(t, account)

	// Get the created entry
	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.Equal(t, entry1.Amount, entry2.Amount)
	require.WithinDuration(t, entry1.CreatedAt, entry2.CreatedAt, time.Second)
}

// TestListEntries lists entries and verifies the returned entries
func TestListEntries(t *testing.T) {
	// Create a random account
	account := createRandomAccount(t)

	// Create 10 random entries for the account
	for i := 0; i < 10; i++ {
		createRandomEntry(t, account)
	}

	// Skip the first 5 entries and get the next 5 entries
	arg := ListEntriesParams{
		AccountID: account.ID,
		Limit:     5,
		Offset:    5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
		require.Equal(t, arg.AccountID, entry.AccountID)
	}
}
