package api

import (
	"encoding/json"
	"errors"
	"github.com/Glide28/go_hw_06/ledger"
	"log"
	"net/http"
	"time"
)

type Server struct{ ledger *ledger.Ledger }

func NewHandler(service *ledger.Ledger) http.Handler {
	s := &Server{ledger: service}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/transactions", s.createTransaction)
	mux.HandleFunc("GET /api/transactions", s.listTransactions)
	mux.HandleFunc("POST /api/budgets", s.createBudget)
	mux.HandleFunc("GET /api/budgets", s.listBudgets)
	return loggingMiddleware(jsonMiddleware(mux))
}

func (s *Server) createTransaction(w http.ResponseWriter, r *http.Request) {
	var request CreateTransactionRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	tx, err := request.toDomain()
	if err != nil {
		writeError(w, http.StatusBadRequest, "date must use YYYY-MM-DD format")
		return
	}
	if err := tx.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	created, err := s.ledger.AddTransaction(tx)
	if errors.Is(err, ledger.ErrBudgetExceeded) {
		writeError(w, http.StatusConflict, "budget exceeded")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, transactionResponse(created))
}

func (s *Server) listTransactions(w http.ResponseWriter, _ *http.Request) {
	items := s.ledger.ListTransactions()
	response := make([]TransactionResponse, 0, len(items))
	for _, tx := range items {
		response = append(response, transactionResponse(tx))
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) createBudget(w http.ResponseWriter, r *http.Request) {
	var request CreateBudgetRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	b := ledger.Budget{Category: request.Category, Limit: request.Limit}
	if err := b.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	created, err := s.ledger.SetBudget(b)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, budgetResponse(created))
}

func (s *Server) listBudgets(w http.ResponseWriter, _ *http.Request) {
	items := s.ledger.ListBudgets()
	response := make([]BudgetResponse, 0, len(items))
	for _, b := range items {
		response = append(response, budgetResponse(b))
	}
	writeJSON(w, http.StatusOK, response)
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("method=%s path=%s duration=%s", r.Method, r.URL.Path, time.Since(started))
	})
}
