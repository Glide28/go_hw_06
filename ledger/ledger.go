package ledger

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var ErrBudgetExceeded = errors.New("budget exceeded")

type Validatable interface{ Validate() error }

type Transaction struct {
	ID          int       `json:"id"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

func (tx Transaction) Validate() error {
	if tx.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	if strings.TrimSpace(tx.Category) == "" {
		return errors.New("category must not be empty")
	}
	if tx.Date.IsZero() {
		return errors.New("date must not be empty")
	}
	return nil
}

type Budget struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
}

func (b Budget) Validate() error {
	if strings.TrimSpace(b.Category) == "" {
		return errors.New("category must not be empty")
	}
	if b.Limit <= 0 {
		return errors.New("limit must be positive")
	}
	return nil
}

type Ledger struct {
	mu           sync.RWMutex
	transactions []Transaction
	budgets      map[string]Budget
	nextID       int
}

func New() *Ledger {
	return &Ledger{transactions: make([]Transaction, 0), budgets: make(map[string]Budget), nextID: 1}
}

func (l *Ledger) AddTransaction(tx Transaction) (Transaction, error) {
	tx.Category = strings.TrimSpace(tx.Category)
	if err := tx.Validate(); err != nil {
		return Transaction{}, fmt.Errorf("invalid transaction: %w", err)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if budget, ok := l.budgets[tx.Category]; ok {
		var spent float64
		for _, saved := range l.transactions {
			if saved.Category == tx.Category {
				spent += saved.Amount
			}
		}
		if spent+tx.Amount > budget.Limit {
			return Transaction{}, ErrBudgetExceeded
		}
	}
	tx.ID = l.nextID
	l.nextID++
	l.transactions = append(l.transactions, tx)
	return tx, nil
}

func (l *Ledger) ListTransactions() []Transaction {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]Transaction, len(l.transactions))
	copy(result, l.transactions)
	return result
}

func (l *Ledger) SetBudget(b Budget) (Budget, error) {
	b.Category = strings.TrimSpace(b.Category)
	if err := b.Validate(); err != nil {
		return Budget{}, fmt.Errorf("invalid budget: %w", err)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.budgets[b.Category] = b
	return b, nil
}

func (l *Ledger) ListBudgets() []Budget {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]Budget, 0, len(l.budgets))
	for _, budget := range l.budgets {
		result = append(result, budget)
	}
	return result
}

// Reset очищает состояние хранилища и возвращает счётчик идентификаторов
// в начальное состояние. Метод используется для изоляции тестовых сценариев.
func (l *Ledger) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.transactions = make([]Transaction, 0)
	l.budgets = make(map[string]Budget)
	l.nextID = 1
}
