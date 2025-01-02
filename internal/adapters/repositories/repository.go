package repositories

import (
	"context"

	"github.com/aniladanir/ethereum-blockchain-parser/internal/core/domain"
)

// Repository defines the interface for accessing transaction data.
type Repository interface {
	// AddTransaction writes transaction to the given address
	AddTransaction(ctx context.Context, address string, transaction domain.Transaction) error

	// GetTransactions returns a list of transactions
	GetTransactions(ctx context.Context, address string) ([]domain.Transaction, error)

	// SetBlockNumber sets the block number
	SetBlockNumber(ctx context.Context, blockNumber int) error

	// GetBlockNumber gets the current block number.
	GetBlockNumber(ctx context.Context) (int, error)

	// AddAddress add the given address to repository
	AddAddress(ctx context.Context, address string) error

	// GetAddresses returns the list of addresses
	GetAddresses(ctx context.Context) ([]string, error)

	// NewTransaction creates a new transaction
	NewTransaction(ctx context.Context) (Transaction, error)
}

type Transaction interface {
	Repository
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
