package api

import (
	"encoding/json"
	"github.com/Glide28/go_hw_06/ledger"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func perform(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func assertJSONContentType(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
}

func TestBudgetEndpoints(t *testing.T) {
	service := ledger.New()
	t.Cleanup(service.Reset)
	handler := NewHandler(service)

	t.Run("create and list", func(t *testing.T) {
		created := perform(handler, http.MethodPost, "/api/budgets", `{"category":"еда","limit":5000}`)
		if created.Code != http.StatusCreated {
			t.Fatalf("POST status = %d, body = %s", created.Code, created.Body.String())
		}
		assertJSONContentType(t, created)

		list := perform(handler, http.MethodGet, "/api/budgets", "")
		if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"category":"еда"`) {
			t.Fatalf("GET status = %d, body = %s", list.Code, list.Body.String())
		}
		assertJSONContentType(t, list)
	})

	t.Run("invalid limit", func(t *testing.T) {
		response := perform(handler, http.MethodPost, "/api/budgets", `{"category":"еда","limit":0}`)
		if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"error"`) {
			t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
		}
		assertJSONContentType(t, response)
	})

	t.Run("bad json", func(t *testing.T) {
		response := perform(handler, http.MethodPost, "/api/budgets", `{"category":`)
		if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"error"`) {
			t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
		}
	})
}

func TestTransactionWorkflow(t *testing.T) {
	service := ledger.New()
	t.Cleanup(service.Reset)
	handler := NewHandler(service)

	t.Run("setup budget", func(t *testing.T) {
		response := perform(handler, http.MethodPost, "/api/budgets", `{"category":"еда","limit":1000}`)
		if response.Code != http.StatusCreated {
			t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
		}
	})

	t.Run("ok", func(t *testing.T) {
		response := perform(handler, http.MethodPost, "/api/transactions", `{"amount":450,"category":"еда","description":"ланч","date":"2026-10-20"}`)
		if response.Code != http.StatusCreated {
			t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
		}
		assertJSONContentType(t, response)
		var tx TransactionResponse
		if err := json.Unmarshal(response.Body.Bytes(), &tx); err != nil || tx.ID == 0 {
			t.Fatalf("invalid response: %s, error = %v", response.Body.String(), err)
		}
	})

	t.Run("list", func(t *testing.T) {
		response := perform(handler, http.MethodGet, "/api/transactions", "")
		var items []TransactionResponse
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d", response.Code)
		}
		if err := json.Unmarshal(response.Body.Bytes(), &items); err != nil || len(items) != 1 {
			t.Fatalf("body = %s, error = %v", response.Body.String(), err)
		}
	})

	t.Run("exceeded", func(t *testing.T) {
		response := perform(handler, http.MethodPost, "/api/transactions", `{"amount":600,"category":"еда","description":"ужин","date":"2026-10-20"}`)
		if response.Code != http.StatusConflict || strings.TrimSpace(response.Body.String()) != `{"error":"budget exceeded"}` {
			t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
		}
		if got := len(service.ListTransactions()); got != 1 {
			t.Fatalf("rejected transaction was saved, count = %d", got)
		}
	})

	t.Run("bad json", func(t *testing.T) {
		response := perform(handler, http.MethodPost, "/api/transactions", `{"amount":`)
		if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"error"`) {
			t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
		}
	})
}
