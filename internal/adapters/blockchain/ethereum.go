package blockchain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aniladanir/ethereum-blockchain-parser/internal/core/domain"
)

// ethereum rpc methods
const (
	ethBlockNumber          = "eth_blockNumber"
	ethGetBlockByNumber     = "eth_getBlockByNumber"
	ethGetTransactionByHash = "eth_getTransactionByHash"
)

const EthereumRpcUrl = "https://ethereum-rpc.publicnode.com"

type ethereumClient struct {
	http.Client
}

func NewEthereumClient(timeout time.Duration) *ethereumClient {
	return &ethereumClient{
		Client: http.Client{
			Timeout: timeout,
		},
	}
}

func (ec *ethereumClient) FetchCurrentBlock(ctx context.Context) (int, error) {
	type responsePayload struct {
		ID      int    `json:"id"`
		JsonRpc string `json:"jsonrpc"`
		Result  string `json:"result"`
	}

	rpcReq := getRpcRequest()
	defer putRpcRequest(rpcReq)

	rpcReq.JsonRpc = "2.0"
	rpcReq.ID = 1
	rpcReq.Method = ethBlockNumber

	body, err := ec.makeRequest(ctx, rpcReq)
	if err != nil {
		return 0, err
	}

	respPayload := new(responsePayload)
	if err := json.Unmarshal(body, respPayload); err != nil {
		return 0, fmt.Errorf("error deserializing response body: %w", err)
	}

	blockNumberStr := strings.Replace(respPayload.Result, "0x", "", 1)
	blockNumber, err := strconv.ParseInt(blockNumberStr, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing block number: %w", err)
	}

	return int(blockNumber), nil
}

func (ec *ethereumClient) FetchBlockByNumber(ctx context.Context, blockNumber int) (*domain.Block, error) {
	type responsePayload struct {
		ID      int           `json:"id"`
		JsonRpc string        `json:"jsonrpc"`
		Result  blockResponse `json:"result"`
	}

	rpcReq := getRpcRequest()
	defer putRpcRequest(rpcReq)

	rpcReq.JsonRpc = "2.0"
	rpcReq.ID = 1
	rpcReq.Method = ethGetBlockByNumber
	rpcReq.Params = append(rpcReq.Params, fmt.Sprintf("0x%x", blockNumber), true)

	body, err := ec.makeRequest(ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	respPayload := new(responsePayload)
	if err := json.Unmarshal(body, respPayload); err != nil {
		return nil, fmt.Errorf("error deserializing response body: %w", err)
	}

	return respPayload.Result.toDomain(), nil
}

func (ec *ethereumClient) makeRequest(ctx context.Context, reqPayload *rpcRequest) ([]byte, error) {
	payloadBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("could not serialize request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, EthereumRpcUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("encountered error when constructing request: %w", err)
	}

	resp, err := ec.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not make request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	return body, nil
}
