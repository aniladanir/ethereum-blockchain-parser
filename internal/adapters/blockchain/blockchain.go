package blockchain

import (
	"context"
	"sync"

	"github.com/aniladanir/ethereum-blockchain-parser/internal/core/domain"
)

// Client represents client for blockchain networks
type Client interface {
	FetchCurrentBlock(ctx context.Context) (int, error)
	FetchBlockByNumber(ctx context.Context, blockNumber int) (*domain.Block, error)
}

type rpcRequest struct {
	ID      int    `json:"id"`
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type blockResponse struct {
	Number           string                `json:"number"`
	Hash             string                `json:"hash"`
	ParentHash       string                `json:"parentHash"`
	Nonce            string                `json:"nonce"`
	Sha3Uncles       string                `json:"sha3Uncles"`
	LogsBloom        string                `json:"logsBloom"`
	TransactionsRoot string                `json:"transactionsRoot"`
	StateRoot        string                `json:"stateRoot"`
	ReceiptsRoot     string                `json:"receiptsRoot"`
	Miner            string                `json:"miner"`
	Difficulty       string                `json:"difficulty"`
	TotalDifficulty  string                `json:"totalDifficulty"`
	ExtraData        string                `json:"extraData"`
	Size             string                `json:"size"`
	GasLimit         string                `json:"gasLimit"`
	GasUsed          string                `json:"gasUsed"`
	Timestamp        string                `json:"timestamp"`
	Transactions     []transactionResponse `json:"transactions"`
	Uncles           []string              `json:"uncles"`
	BaseFeePerGas    string                `json:"baseFeePerGas"`
}

func (b *blockResponse) toDomain() *domain.Block {
	domainBlock := &domain.Block{
		Number: b.Number,
	}

	for i := range b.Transactions {
		domainBlock.Transactions = append(domainBlock.Transactions, domain.Transaction{
			Hash:        b.Transactions[i].Hash,
			From:        b.Transactions[i].From,
			To:          b.Transactions[i].To,
			BlockNumber: b.Transactions[i].BlockNumber,
			Value:       b.Transactions[i].Value,
		})
	}

	return domainBlock
}

type transactionResponse struct {
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	From             string `json:"from"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	Nonce            string `json:"nonce"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
	Value            string `json:"value"`
	V                string `json:"v"`
	R                string `json:"r"`
	S                string `json:"s"`
}

func (t *transactionResponse) toDomain() *domain.Transaction {
	return &domain.Transaction{
		Hash:        t.Hash,
		From:        t.From,
		To:          t.To,
		Value:       t.Value,
		BlockNumber: t.BlockNumber,
	}
}

var (
	rpcReqPool = sync.Pool{
		New: func() any {
			return new(rpcRequest)
		},
	}
)

func getRpcRequest() *rpcRequest {
	return rpcReqPool.Get().(*rpcRequest)
}

func putRpcRequest(rpc *rpcRequest) {
	if rpc == nil {
		return
	}

	// zero object
	rpc.ID = 0
	rpc.JsonRpc = ""
	rpc.Method = ""
	rpc.Params = rpc.Params[:0]

	rpcReqPool.Put(rpc)
}
