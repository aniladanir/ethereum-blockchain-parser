package httphandler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/aniladanir/ethereum-blockchain-parser/internal/core/domain"
	"github.com/aniladanir/ethereum-blockchain-parser/internal/core/services"
	"github.com/aniladanir/ethereum-blockchain-parser/pkg/errs"
)

type HttpHandler struct {
	server   http.Server
	txParser services.TransactionParser
	logger   *slog.Logger
}

type Response struct {
	Msg  string `json:"msg,omitempty"`
	Data any    `json:"data,omitempty"`
}

func NewHttpHandler(addr string, txParser services.TransactionParser, logger *slog.Logger) *HttpHandler {
	httpHandler := &HttpHandler{
		server: http.Server{
			Addr:              addr,
			ReadTimeout:       time.Second * 5,
			WriteTimeout:      time.Second * 5,
			IdleTimeout:       time.Second * 5,
			ReadHeaderTimeout: time.Second * 5,
		},
		txParser: txParser,
		logger:   logger,
	}

	// register handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/api/block", httpHandler.getCurrentBlockNumber)
	mux.HandleFunc("/api/subscribe", httpHandler.subscribeToAddress)
	mux.HandleFunc("/api/transactions", httpHandler.getTransactionsByAddress)
	httpHandler.server.Handler = mux

	return httpHandler
}

func (h *HttpHandler) Listen() error {
	return h.server.ListenAndServe()
}

func (h *HttpHandler) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func (h *HttpHandler) getCurrentBlockNumber(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// set content type
	w.Header().Set("Content-Type", "application/json")

	// get last parsed block
	currentBlock, err := h.txParser.GetCurrentBlock(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		h.logger.Error(err.Error())
		return
	}

	// write to response body
	err = json.NewEncoder(w).Encode(&Response{
		Msg: "success",
		Data: struct {
			CurrentBlock int `json:"currentBlock"`
		}{
			CurrentBlock: currentBlock,
		},
	})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		h.logger.Error("error writing to response body", slog.Any("error", err))
	}
}

func (h *HttpHandler) subscribeToAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// set content type
	w.Header().Set("Content-Type", "application/json")

	// get query params
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address query param is required", http.StatusBadRequest)
		h.logger.Error("missing address query param")
		return
	}

	// subscribe to the provided address
	if err := h.txParser.Subscribe(r.Context(), address); err != nil {
		if errs.IsAlreadyExistErr(err) {
			http.Error(w, "provided address is already subscribed", http.StatusConflict)
			h.logger.Error(err.Error())
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			h.logger.Error(err.Error())
		}
		return
	}

	// write to response body
	err := json.NewEncoder(w).Encode(&Response{
		Msg: "success",
	})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		h.logger.Error("error writing to response body", slog.Any("error", err))
	}
}

func (h *HttpHandler) getTransactionsByAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// set content type
	w.Header().Set("Content-Type", "application/json")

	// get query params
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address query param is required", http.StatusBadRequest)
		h.logger.Error("missing address query param")
		return
	}

	// get transactions belonging to the given address
	transactions, err := h.txParser.GetTransactions(r.Context(), address)
	if err != nil {
		if errs.IsNotFoundErr(err) {
			http.Error(w, "the address does not exist in our records", http.StatusNotFound)
			h.logger.Error(err.Error())
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			h.logger.Error(err.Error())
		}
		return
	}

	// write to response body
	err = json.NewEncoder(w).Encode(&Response{
		Msg: "success",
		Data: struct {
			Transactions []domain.Transaction `json:"transactions"`
		}{
			Transactions: transactions,
		},
	})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		h.logger.Error("error writing to response body", slog.Any("error", err))
	}
}
