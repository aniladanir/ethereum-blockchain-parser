package httphandler

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aniladanir/ethereum-blockchain-parser/internal/core/domain"
	"github.com/aniladanir/ethereum-blockchain-parser/pkg/errs"
)

// Define a mock struct for txParser for testing
type MockTxParser struct {
	currentBlock      int
	transactions      []domain.Transaction
	subscribeError    error
	transactionsError error
}

func (m *MockTxParser) GetCurrentBlock(ctx context.Context) (int, error) {
	return m.currentBlock, nil
}
func (m *MockTxParser) Subscribe(ctx context.Context, address string) error {
	return m.subscribeError
}
func (m *MockTxParser) GetTransactions(ctx context.Context, address string) ([]domain.Transaction, error) {
	return m.transactions, m.transactionsError
}
func (m *MockTxParser) ProcessNewBlocks(ctx context.Context, interval time.Duration) error {
	return nil
}

func setupTest(txParser *MockTxParser) *HttpHandler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return &HttpHandler{txParser: txParser, logger: logger}
}

func TestCurrentBlockHandler(t *testing.T) {
	tests := []struct {
		name           string
		txParser       *MockTxParser
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			txParser:       &MockTxParser{currentBlock: 100},
			expectedStatus: http.StatusOK,
			expectedBody: `{"msg":"success","data":{"currentBlock":100}}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setupTest(tt.txParser)
			req := httptest.NewRequest(http.MethodGet, "/currentblock", nil)
			rec := httptest.NewRecorder()

			h.getCurrentBlockNumber(rec, req)
			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, rec.Code)
			}

			actualBody := rec.Body.String()
			if actualBody != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, actualBody)
			}
		})
	}
}

func TestSubscribeHandler(t *testing.T) {
	tests := []struct {
		name           string
		txParser       *MockTxParser
		address        string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			txParser:       &MockTxParser{},
			address:        "0x123",
			expectedStatus: http.StatusOK,
			expectedBody: `{"msg":"success"}
`,
		},
		{
			name:           "Missing Address",
			txParser:       &MockTxParser{},
			address:        "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "address query param is required\n",
		},
		{
			name:           "Already Subscribed",
			txParser:       &MockTxParser{subscribeError: errs.AlreadyExistErr()},
			address:        "0x123",
			expectedStatus: http.StatusConflict,
			expectedBody:   "provided address is already subscribed\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setupTest(tt.txParser)
			req := httptest.NewRequest(http.MethodPost, "/subscribe?address="+tt.address, nil)

			rec := httptest.NewRecorder()
			h.subscribeToAddress(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, rec.Code)
			}

			actualBody := rec.Body.String()
			if actualBody != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, actualBody)
			}
		})
	}
}

func TestTransactionsHandler(t *testing.T) {

	tests := []struct {
		name           string
		txParser       *MockTxParser
		address        string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			txParser:       &MockTxParser{transactions: []domain.Transaction{{Hash: "hash1", From: "from1", To: "to1", Value: "100", BlockNumber: "1"}}},
			address:        "0x123",
			expectedStatus: http.StatusOK,
			expectedBody: `{"msg":"success","data":{"transactions":[{"hash":"hash1","from":"from1","to":"to1","value":"100","blockNumber":"1"}]}}
`,
		},
		{
			name:           "Missing Address",
			txParser:       &MockTxParser{},
			address:        "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "address query param is required\n",
		},
		{
			name:           "Address Not Found",
			txParser:       &MockTxParser{transactionsError: errs.NotFoundErr()},
			address:        "0x123",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "the address does not exist in our records\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setupTest(tt.txParser)
			req := httptest.NewRequest(http.MethodGet, "/transactions?address="+tt.address, nil)

			rec := httptest.NewRecorder()
			h.getTransactionsByAddress(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, rec.Code)
			}
			actualBody := rec.Body.String()
			if actualBody != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, actualBody)
			}
		})
	}
}
