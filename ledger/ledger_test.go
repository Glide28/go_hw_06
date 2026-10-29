package ledger

import (
	"errors"
	"testing"
	"time"
)

func TestTransactionValidate(t *testing.T) {
	t.Parallel()
	validDate := time.Date(2026, time.October, 20, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		tx      Transaction
		wantErr bool
	}{
		{name: "valid", tx: Transaction{Amount: 100, Category: "еда", Date: validDate}},
		{name: "zero amount", tx: Transaction{Amount: 0, Category: "еда", Date: validDate}, wantErr: true},
		{name: "negative amount", tx: Transaction{Amount: -1, Category: "еда", Date: validDate}, wantErr: true},
		{name: "empty category", tx: Transaction{Amount: 100, Category: "   ", Date: validDate}, wantErr: true},
		{name: "empty date", tx: Transaction{Amount: 100, Category: "еда"}, wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.tx.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBudgetValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		budget  Budget
		wantErr bool
	}{
		{name: "valid", budget: Budget{Category: "еда", Limit: 5000}},
		{name: "zero limit", budget: Budget{Category: "еда", Limit: 0}, wantErr: true},
		{name: "negative limit", budget: Budget{Category: "еда", Limit: -100}, wantErr: true},
		{name: "empty category", budget: Budget{Category: " ", Limit: 5000}, wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.budget.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLedgerBudgetLimit(t *testing.T) {
	l := New()
	t.Cleanup(l.Reset)

	if _, err := l.SetBudget(Budget{Category: "еда", Limit: 5000}); err != nil {
		t.Fatalf("SetBudget() error = %v", err)
	}
	if _, err := l.AddTransaction(Transaction{Amount: 3000, Category: "еда", Date: time.Now()}); err != nil {
		t.Fatalf("first AddTransaction() error = %v", err)
	}
	before := len(l.ListTransactions())
	_, err := l.AddTransaction(Transaction{Amount: 2500, Category: "еда", Date: time.Now()})
	if !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("expected ErrBudgetExceeded, got %v", err)
	}
	if after := len(l.ListTransactions()); after != before {
		t.Fatalf("rejected transaction was saved: before=%d, after=%d", before, after)
	}
}

func TestReset(t *testing.T) {
	l := New()
	_, _ = l.SetBudget(Budget{Category: "еда", Limit: 1000})
	_, _ = l.AddTransaction(Transaction{Amount: 100, Category: "еда", Date: time.Now()})
	l.Reset()

	if len(l.ListBudgets()) != 0 || len(l.ListTransactions()) != 0 {
		t.Fatal("Reset() did not clear ledger")
	}
	created, err := l.AddTransaction(Transaction{Amount: 100, Category: "еда", Date: time.Now()})
	if err != nil || created.ID != 1 {
		t.Fatalf("counter was not reset: transaction=%+v, error=%v", created, err)
	}
}
