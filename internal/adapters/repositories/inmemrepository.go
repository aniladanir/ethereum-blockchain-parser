package repositories

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/aniladanir/ethereum-blockchain-parser/internal/core/domain"
	"github.com/aniladanir/ethereum-blockchain-parser/pkg/errs"
)

var (
	_ Repository  = (*inMemRepository)(nil)
	_ Transaction = (*inMemTransaction)(nil)
)

type inMemRepository struct {
	addresses    *sync.Map
	transactions map[string][]domain.Transaction
	blockNumber  *atomic.Int64
}

type inMemTransaction struct {
	*inMemRepository
}

func NewInmemTransactionRepository() Repository {
	a := &inMemRepository{
		blockNumber: &atomic.Int64{},
		addresses:   new(sync.Map),
		transactions: map[string][]domain.Transaction{
			"0x123": {{
				Hash:        "00000",
				From:        "0x123",
				To:          "0x456",
				Value:       "hello world!",
				BlockNumber: "20000000",
			}},
		},
	}
	a.addresses.Store("0x123", new(sync.RWMutex))
	return a
}

func (tr *inMemRepository) NewTransaction(ctx context.Context) (Transaction, error) {
	return &inMemTransaction{
		inMemRepository: tr,
	}, nil
}

func (tr *inMemRepository) GetBlockNumber(ctx context.Context) (int, error) {
	return int(tr.blockNumber.Load()), nil
}

func (tr *inMemRepository) SetBlockNumber(ctx context.Context, blockNumber int) error {
	tr.blockNumber.Store(int64(blockNumber))
	return nil
}

func (tr *inMemRepository) GetTransactions(ctx context.Context, address string) ([]domain.Transaction, error) {
	// get transactions rw mutex
	transactionsMtxAny, ok := tr.addresses.Load(address)
	if !ok {
		return nil, errs.NotFoundErr()
	}
	transactionsMtx := transactionsMtxAny.(*sync.RWMutex)

	// read-lock transactions mutex
	transactionsMtx.RLock()
	defer transactionsMtx.RUnlock()

	// copy transaction records
	transactions := tr.transactions[address]
	if len(transactions) == 0 {
		return make([]domain.Transaction, 0), nil
	}
	transactionCopy := make([]domain.Transaction, len(transactions))
	_ = copy(transactionCopy, transactions)

	return transactionCopy, nil
}

func (tr *inMemRepository) AddTransaction(ctx context.Context, address string, transaction domain.Transaction) error {
	// get transactions mutex
	transactionsMtxAny, ok := tr.addresses.Load(address)
	if !ok {
		return errs.NotFoundErr()
	}
	transactionsMtx := transactionsMtxAny.(*sync.RWMutex)

	// write-lock transactions mutex
	transactionsMtx.Lock()
	defer transactionsMtx.Unlock()

	// write transaction
	if _, ok = tr.transactions[address]; !ok {
		tr.transactions[address] = make([]domain.Transaction, 0)
	}
	tr.transactions[address] = append(tr.transactions[address], transaction)

	return nil
}

func (tr *inMemRepository) AddAddress(ctx context.Context, address string) error {
	if _, ok := tr.addresses.LoadOrStore(address, new(sync.RWMutex)); ok {
		return errs.AlreadyExistErr()
	}
	return nil
}

func (tr *inMemRepository) GetAddresses(ctx context.Context) ([]string, error) {
	addresses := make([]string, 0)
	tr.addresses.Range(func(address, _ any) bool {
		addresses = append(addresses, address.(string))
		return true
	})
	return addresses, nil
}

func (tr *inMemTransaction) Commit(ctx context.Context) error {
	return nil
}

func (tr *inMemTransaction) Rollback(ctx context.Context) error {
	return nil
}
