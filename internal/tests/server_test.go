package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	httpServer "test-assignment/internal/http"
	"test-assignment/internal/models"
	"test-assignment/mocks"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
)

func TestHandleWithdrawal_Success(t *testing.T) {
	mockStorage := &mocks.Storage{}

	mockStorage.On(
		"Withdrawal",
		mock.Anything,
		1,
		mock.AnythingOfType("models.Withdrawal"),
		"key-1",
	).Return(models.Withdrawal{
		UserID:      1,
		Amount:      "10",
		Currency:    "USDT",
		Destination: "Wallet",
		Status:      "pending",
	}, nil)

	handlers := httpServer.NewHTTPHandlers(mockStorage)
	router := mux.NewRouter()
	router.HandleFunc("/v1/withdrawals", handlers.HandleWithdrawal).Methods("POST")

	reqBody := httpServer.WithdrawalReqDTO{
		UserID:         1,
		Amount:         10,
		Currency:       "USDT",
		Destination:    "Wallet",
		IdempotencyKey: "key-1",
	}

	w := performRequest(router, "POST", "/v1/withdrawals", reqBody)

	if w.Code != 200 {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	mockStorage.AssertExpectations(t)
}

func TestWithdrawal_Idempotency(t *testing.T) {
	mockStorage := &mocks.Storage{}

	withdrawal := models.Withdrawal{
		UserID:      1,
		Amount:      "100",
		Currency:    "USDT",
		Destination: "wallet",
		Status:      "pending",
	}

	mockStorage.On(
		"Withdrawal",
		mock.Anything,
		1,
		mock.AnythingOfType("models.Withdrawal"),
		"idemp-key-1",
	).Return(withdrawal, nil).Once()

	mockStorage.On(
		"Withdrawal",
		mock.Anything,
		1,
		mock.AnythingOfType("models.Withdrawal"),
		"idemp-key-1",
	).Return(withdrawal, nil).Once()

	handlers := httpServer.NewHTTPHandlers(mockStorage)
	router := mux.NewRouter()
	router.HandleFunc("/v1/withdrawals", handlers.HandleWithdrawal).Methods("POST")

	reqBody := httpServer.WithdrawalReqDTO{
		UserID:         1,
		Amount:         100,
		Currency:       "USDT",
		Destination:    "wallet",
		IdempotencyKey: "idemp-key-1",
	}

	w1 := performRequest(router, "POST", "/v1/withdrawals", reqBody)
	if w1.Code != 200 {
		t.Fatalf("expected 200 OK, got %d", w1.Code)
	}

	w2 := performRequest(router, "POST", "/v1/withdrawals", reqBody)
	if w2.Code != 200 {
		t.Fatalf("expected 200 OK for idempotent call, got %d", w2.Code)
	}

	mockStorage.AssertExpectations(t)
}

func TestWithdrawal_IdempotencyConflict(t *testing.T) {
	mockStorage := &mocks.Storage{}

	withdrawal1 := models.Withdrawal{
		UserID:      1,
		Amount:      "100",
		Currency:    "USDT",
		Destination: "wallet",
		Status:      "pending",
	}

	mockStorage.On(
		"Withdrawal",
		mock.Anything,
		1,
		mock.AnythingOfType("models.Withdrawal"),
		"idemp-key-1",
	).Return(withdrawal1, nil)

	handlers := httpServer.NewHTTPHandlers(mockStorage)
	router := mux.NewRouter()
	router.HandleFunc("/v1/withdrawals", handlers.HandleWithdrawal).Methods("POST")

	reqBody1 := httpServer.WithdrawalReqDTO{
		UserID:         1,
		Amount:         100,
		Currency:       "USDT",
		Destination:    "wallet",
		IdempotencyKey: "idemp-key-1",
	}
	w1 := performRequest(router, "POST", "/v1/withdrawals", reqBody1)
	if w1.Code != 200 {
		t.Fatalf("expected 200 OK, got %d", w1.Code)
	}

	reqBody2 := httpServer.WithdrawalReqDTO{
		UserID:         1,
		Amount:         200,
		Currency:       "USDT",
		Destination:    "wallet2",
		IdempotencyKey: "idemp-key-1",
	}
	w2 := performRequest(router, "POST", "/v1/withdrawals", reqBody2)
	if w2.Code != 422 {
		t.Fatalf("expected 422 for idempotency conflict, got %d", w2.Code)
	}

	mockStorage.AssertExpectations(t)
}

func TestConcurrentWithdrawals(t *testing.T) {
	mockStorage := &mocks.Storage{}

	mockStorage.On(
		"Withdrawal",
		mock.Anything,
		1,
		mock.AnythingOfType("models.Withdrawal"),
		mock.AnythingOfType("string"),
	).Return(func(ctx context.Context, userID int, w models.Withdrawal, key string) models.Withdrawal {
		return models.Withdrawal{
			UserID:      userID,
			Amount:      w.Amount,
			Currency:    w.Currency,
			Destination: w.Destination,
			Status:      "pending",
		}
	}, nil).Times(5)

	handlers := httpServer.NewHTTPHandlers(mockStorage)
	router := mux.NewRouter()
	router.HandleFunc("/v1/withdrawals", handlers.HandleWithdrawal).Methods("POST")

	const goroutines = 5
	wg := &sync.WaitGroup{}
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()

			reqBody := httpServer.WithdrawalReqDTO{
				UserID:         1,
				Amount:         10 + i,
				Currency:       "USDT",
				Destination:    fmt.Sprintf("wallet-%d", i),
				IdempotencyKey: fmt.Sprintf("key-%d", i),
			}

			w := performRequest(router, "POST", "/v1/withdrawals", reqBody)
			if w.Code != 200 {
				t.Errorf("expected 200 OK, got %d for request %d", w.Code, i)
			}
		}(i)
	}

	wg.Wait()
	mockStorage.AssertExpectations(t)
}

func performRequest(handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}
