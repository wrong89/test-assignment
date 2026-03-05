package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"test-assignment/internal/models"
	"test-assignment/internal/storage"
	"test-assignment/internal/storage/repo"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

type HTTPHandlers struct {
	storage repo.Storage
}

func NewHTTPHandlers(storage repo.Storage) *HTTPHandlers {
	return &HTTPHandlers{
		storage: storage,
	}
}

func (h *HTTPHandlers) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var createUserReqDTO CreateUserReqDTO

	if err := json.NewDecoder(r.Body).Decode(&createUserReqDTO); err != nil {
		errDTO := NewErrorDTO(err)
		http.Error(w, errDTO.String(), http.StatusBadRequest)
		return
	}

	if err := createUserReqDTO.Validate(); err != nil {
		errDTO := NewErrorDTO(err)
		http.Error(w, errDTO.String(), http.StatusBadRequest)
		return
	}

	decimal := decimal.NewFromInt(createUserReqDTO.Balance)

	id, err := h.storage.CreateUser(r.Context(), createUserReqDTO.UserID, decimal)
	if err != nil {
		errDTO := NewErrorDTO(err)
		if errors.Is(err, storage.ErrUserExist) {
			http.Error(w, errDTO.String(), http.StatusConflict)
			return
		}

		http.Error(w, errDTO.String(), http.StatusInternalServerError)
		return
	}

	resp := CreateUserResDTO{UserID: id}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *HTTPHandlers) HandleWithdrawal(w http.ResponseWriter, r *http.Request) {
	var withdrawalReqDTO WithdrawalReqDTO

	if err := json.NewDecoder(r.Body).Decode(&withdrawalReqDTO); err != nil {
		errDTO := NewErrorDTO(err)
		http.Error(w, errDTO.String(), http.StatusBadRequest)
		return
	}

	if err := withdrawalReqDTO.Validate(); err != nil {
		errDTO := NewErrorDTO(err)
		http.Error(w, errDTO.String(), http.StatusBadRequest)
		return
	}

	mwd := models.Withdrawal{
		UserID:      withdrawalReqDTO.UserID,
		Amount:      decimal.NewFromInt(int64(withdrawalReqDTO.Amount)).String(),
		Currency:    withdrawalReqDTO.Currency,
		Destination: withdrawalReqDTO.Destination,
		Status:      "pending",
	}

	wd, err := h.storage.Withdrawal(
		r.Context(),
		withdrawalReqDTO.UserID,
		mwd,
		withdrawalReqDTO.IdempotencyKey,
	)
	if err != nil {
		errDTO := NewErrorDTO(err)
		if errors.Is(err, storage.ErrInsufficientBalance) {
			http.Error(w, errDTO.String(), http.StatusConflict)
			return
		}

		http.Error(w, errDTO.String(), http.StatusInternalServerError)
		return
	}

	if sameWithdrawal(wd, mwd) {
		json.NewEncoder(w).Encode(wd)
		return
	}

	w.WriteHeader(http.StatusUnprocessableEntity)
}

func (h *HTTPHandlers) HandleGetWithdrawal(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		errDTO := NewErrorDTO(ErrWithdrawalIDEmpty)
		http.Error(w, errDTO.String(), http.StatusBadRequest)
		return
	}

	idNum, err := strconv.Atoi(id)
	if err != nil {
		errDTO := NewErrorDTO(ErrBadIDFormat)
		http.Error(w, errDTO.String(), http.StatusBadRequest)
		return
	}

	wd, err := h.storage.GetWithdrawal(r.Context(), idNum)
	if err != nil {
		errDTO := NewErrorDTO(err)

		if errors.Is(err, storage.ErrWithdrawalNotExist) {
			http.Error(w, errDTO.String(), http.StatusNotFound)
			return
		}

		http.Error(w, errDTO.String(), http.StatusInternalServerError)
		return
	}

	res := GetWithdrawalRes{
		ID:          wd.ID,
		UserID:      wd.UserID,
		Amount:      wd.Amount,
		Currency:    wd.Currency,
		Destination: wd.Destination,
	}

	json.NewEncoder(w).Encode(res)
}
