package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aniladanir/ethereum-blockchain-parser/internal/adapters/blockchain"
	"github.com/aniladanir/ethereum-blockchain-parser/internal/adapters/repositories"
	"github.com/aniladanir/ethereum-blockchain-parser/internal/core/domain"
)

type TransactionParser interface {
	// ProcessNewBlocks is a blocking function that continuously searches for newly mined blocks on the blockchain network
	// that have not been processed.
	//
	// 'internal' argument determines the duration between each attempt
	ProcessNewBlocks(ctx context.Context, interval time.Duration) error

	// GetCurrentBlock returns the last parsed block in the blockchain
	GetCurrentBlock(ctx context.Context) (int, error)

	// Subscribe adds the given address to be observed by the transaction service
	Subscribe(ctx context.Context, address string) error

	// GetTransactions returns a list of inbound and outbound transactions for a given address
	GetTransactions(ctx context.Context, address string) ([]domain.Transaction, error)
}

var _ TransactionParser = (*transactionParser)(nil)

type transactionParser struct {
	logger   *slog.Logger
	bcClient blockchain.Client
	repo     repositories.Repository
}

func NewTransactionParser(repo repositories.Repository, bcClient blockchain.Client, logger *slog.Logger) TransactionParser {
	return &transactionParser{
		logger:   logger,
		bcClient: bcClient,
		repo:     repo,
	}
}

func (tp *transactionParser) GetCurrentBlock(ctx context.Context) (int, error) {
	return tp.repo.GetBlockNumber(ctx)
}

func (tp *transactionParser) Subscribe(ctx context.Context, address string) error {
	return tp.repo.AddAddress(ctx, address)
}

func (tp *transactionParser) GetTransactions(ctx context.Context, address string) ([]domain.Transaction, error) {
	return tp.repo.GetTransactions(ctx, address)
}

// ProcessNewBlocks is a blocking function that continuously searches for newly mined blocks on the blockchain network
// that have not been processed.
//
// 'interval' argument determines the duration between each process cycle
func (tp *transactionParser) ProcessNewBlocks(ctx context.Context, interval time.Duration) error {
	if err := tp.updateBlockNumber(ctx); err != nil {
		return err
	}
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			tp.processNewBlocks(ctx)
		}
	}
}

func (tp *transactionParser) processNewBlocks(ctx context.Context) {
	// fetch most recently mined block number from blockchain
	lastMinedBlock, err := tp.bcClient.FetchCurrentBlock(ctx)
	if err != nil {
		tp.logger.Error("could not fetch current block number", slog.Any("error", err))
		return
	}

	// get last fetched block number
	lastProcessedBlock, err := tp.GetCurrentBlock(ctx)
	if err != nil {
		tp.logger.Error("could not get current block number from repository", slog.Any("error", err))
		return
	}

	// compare last fetched block number to latest mined block
	// if there is no difference, do nothing
	if lastMinedBlock <= lastProcessedBlock {
		return
	}

	// get subscribed addresses
	subscribedAddresses, err := tp.repo.GetAddresses(ctx)
	if err != nil {
		tp.logger.Error("could not get subscribed addresses from repository", slog.Any("error", err))
		return
	}

	// create new repository transaction
	repoTx, err := tp.repo.NewTransaction(ctx)
	if err != nil {
		tp.logger.Error("could not create repository transaction", slog.Any("error", err))
		return
	}

	// catch up to the last fetched block number
	for block := lastProcessedBlock + 1; block <= lastMinedBlock; block++ {
		blockData, err := tp.bcClient.FetchBlockByNumber(ctx, block)
		if err != nil {
			tp.logger.Error("could not fetch block", slog.Any("error", err), slog.Int("block number", block))
			return
		}

		// process transactions
		for i := range blockData.Transactions {
			// if any transaction is outgoing or incoming to the one of the addresses in
			// the subscription list, add it to the repository
			for _, addr := range subscribedAddresses {
				if blockData.Transactions[i].From == addr || blockData.Transactions[i].To == addr {
					if err := repoTx.AddTransaction(ctx, addr, blockData.Transactions[i]); err != nil {
						tp.logger.Error("could not add transaction to the repository", slog.Any("error", err))
						if err := repoTx.Rollback(ctx); err != nil {
							tp.logger.Error("could not rollback repository transaction", slog.Any("error", err))
						}
						return
					}
				}
			}
		}

		// set the processed block number in repository
		if err := repoTx.SetBlockNumber(ctx, block); err != nil {
			tp.logger.Error("could not set block number in repository", slog.Any("error", err), slog.Int("block number", block))
			if err := repoTx.Rollback(ctx); err != nil {
				tp.logger.Error("could not rollback repository transaction", slog.Any("error", err))
			}
			return
		}
	}

	// commit transaction
	if err := repoTx.Commit(ctx); err != nil {
		tp.logger.Error("could not commit repository transaction", slog.Any("error", err))
	}
}

func (tp *transactionParser) updateBlockNumber(ctx context.Context) error {
	// set current block number
	blockNumber, err := tp.bcClient.FetchCurrentBlock(ctx)
	if err != nil {
		return fmt.Errorf("could not fetch current block number: %w", err)
	}
	if err = tp.repo.SetBlockNumber(ctx, blockNumber); err != nil {
		return fmt.Errorf("could not set block number: %w", err)
	}
	return nil
}
