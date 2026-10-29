package api

import (
	"github.com/Glide28/go_hw_06/ledger"
	"time"
)

type CreateTransactionRequest struct {
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}
type TransactionResponse struct {
	ID          int     `json:"id"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}
type CreateBudgetRequest struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
}
type BudgetResponse struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
}

func (r CreateTransactionRequest) toDomain() (ledger.Transaction, error) {
	date, err := time.Parse("2006-01-02", r.Date)
	if err != nil {
		return ledger.Transaction{}, err
	}
	return ledger.Transaction{Amount: r.Amount, Category: r.Category, Description: r.Description, Date: date}, nil
}
func transactionResponse(tx ledger.Transaction) TransactionResponse {
	return TransactionResponse{ID: tx.ID, Amount: tx.Amount, Category: tx.Category, Description: tx.Description, Date: tx.Date.Format("2006-01-02")}
}
func budgetResponse(b ledger.Budget) BudgetResponse {
	return BudgetResponse{Category: b.Category, Limit: b.Limit}
}
