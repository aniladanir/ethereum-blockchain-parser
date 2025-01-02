package repositories

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/aniladanir/ethereum-blockchain-parser/internal/core/domain"
	"github.com/aniladanir/ethereum-blockchain-parser/pkg/errs"
)

func TestGetTransactions(t *testing.T) {
	tests := []struct {
		name           string
		initialState   map[string][]domain.Transaction
		address        string
		expectedResult []domain.Transaction
		expectedError  error
	}{
		{
			name: "Success",
			initialState: map[string][]domain.Transaction{
				"0x123": {{Hash: "hash1", From: "from1", To: "to1", Value: "100", BlockNumber: "1"}},
			},
			address:        "0x123",
			expectedResult: []domain.Transaction{{Hash: "hash1", From: "from1", To: "to1", Value: "100", BlockNumber: "1"}},
			expectedError:  nil,
		},
		{
			name:           "AddressNotFound",
			initialState:   map[string][]domain.Transaction{},
			address:        "0x456",
			expectedResult: nil,
			expectedError:  errs.NotFoundErr(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTest(tt.initialState, map[string]any{"0x123": new(sync.RWMutex)})
			transactions, err := repo.GetTransactions(context.Background(), tt.address)

			if !errors.Is(err, tt.expectedError) {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
			if len(transactions) != len(tt.expectedResult) {
				t.Errorf("expected %d transactions got %d", len(tt.expectedResult), len(transactions))
				return
			}

			for i := range transactions {
				if transactions[i] != tt.expectedResult[i] {
					t.Errorf("expected transaction %v, got %v", tt.expectedResult[i], transactions[i])
				}
			}

		})
	}
}

func TestAddTransaction(t *testing.T) {
	tests := []struct {
		name             string
		initialState     map[string][]domain.Transaction
		address          string
		transaction      domain.Transaction
		expectedError    error
		initialAddresses map[string]any
	}{
		{
			name:             "Success",
			initialState:     map[string][]domain.Transaction{},
			address:          "0x123",
			transaction:      domain.Transaction{Hash: "hash1", From: "from1", To: "to1", Value: "100", BlockNumber: "1"},
			expectedError:    nil,
			initialAddresses: map[string]any{"0x123": new(sync.RWMutex)},
		},
		{
			name:             "AddressNotFound",
			initialState:     map[string][]domain.Transaction{},
			address:          "0x456",
			transaction:      domain.Transaction{Hash: "hash1", From: "from1", To: "to1", Value: "100", BlockNumber: "1"},
			expectedError:    errs.NotFoundErr(),
			initialAddresses: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTest(tt.initialState, tt.initialAddresses)
			err := repo.AddTransaction(context.Background(), tt.address, tt.transaction)
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
			if tt.expectedError == nil {
				transactions, _ := repo.GetTransactions(context.Background(), tt.address)
				if len(transactions) != 1 {
					t.Errorf("expected to add one transaction, and found %d", len(transactions))
				}

				if transactions[0] != tt.transaction {
					t.Errorf("expected the transaction to be %v and found %v", tt.transaction, transactions[0])
				}
			}

		})
	}
}

func TestAddAddress(t *testing.T) {
	tests := []struct {
		name          string
		initialState  map[string]any
		address       string
		expectedError error
	}{
		{
			name:          "Success",
			initialState:  map[string]any{},
			address:       "0x123",
			expectedError: nil,
		},
		{
			name:          "AlreadyExist",
			initialState:  map[string]any{"0x123": new(sync.RWMutex)},
			address:       "0x123",
			expectedError: errs.AlreadyExistErr(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTest(nil, tt.initialState)
			err := repo.AddAddress(context.Background(), tt.address)
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestGetAddresses(t *testing.T) {
	tests := []struct {
		name           string
		initialState   map[string]any
		expectedResult []string
		expectedError  error
	}{
		{
			name: "Success",
			initialState: map[string]any{
				"0x123": new(sync.RWMutex),
				"0x456": new(sync.RWMutex),
			},
			expectedResult: []string{"0x123", "0x456"},
			expectedError:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTest(nil, tt.initialState)
			addresses, err := repo.GetAddresses(context.Background())

			if !errors.Is(err, tt.expectedError) {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}

			if len(addresses) != len(tt.expectedResult) {
				t.Errorf("expected to return %d address and returned %d", len(tt.expectedResult), len(addresses))
				return
			}

			if addresses[0] != tt.expectedResult[0] {
				t.Errorf("expected address to be %s and found %s", tt.expectedResult[0], addresses[0])
			}
			if addresses[1] != tt.expectedResult[1] {
				t.Errorf("expected address to be %s and found %s", tt.expectedResult[1], addresses[1])
			}
		})
	}
}

func setupTest(initialTransactions map[string][]domain.Transaction, initialAddresses map[string]any) *inMemRepository {
	addresses := new(sync.Map)
	for key, value := range initialAddresses {
		addresses.Store(key, value)
	}
	return &inMemRepository{
		transactions: initialTransactions,
		addresses:    addresses,
	}
}
