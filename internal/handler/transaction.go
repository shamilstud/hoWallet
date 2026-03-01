package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/howallet/howallet/internal/middleware"
	"github.com/howallet/howallet/internal/model"
	"github.com/howallet/howallet/internal/service"
)

type TransactionHandler struct {
	txnSvc *service.TransactionService
}

func NewTransactionHandler(txnSvc *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{txnSvc: txnSvc}
}

// POST /api/transactions
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTransactionRequest
	if err := Decode(r, &req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Description == "" || req.Amount == "" {
		ErrorJSON(w, http.StatusBadRequest, "description and amount are required")
		return
	}

	userID := middleware.UserIDFromCtx(r.Context())
	hhID := middleware.HouseholdIDFromCtx(r.Context())

	txn, err := h.txnSvc.Create(r.Context(), hhID, userID, req)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	JSON(w, http.StatusCreated, txn)
}

// GET /api/transactions
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	hhID := middleware.HouseholdIDFromCtx(r.Context())

	q := model.ListTransactionsQuery{
		Limit:  50,
		Offset: 0,
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			q.Limit = int32(n)
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			q.Offset = int32(n)
		}
	}
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q.From = &t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q.To = &t
		}
	}
	if v := r.URL.Query().Get("type"); v != "" {
		tt := model.TransactionType(v)
		q.Type = &tt
	}
	if v := r.URL.Query().Get("account_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			q.AccountID = &id
		}
	}

	result, err := h.txnSvc.List(r.Context(), hhID, q)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to list transactions")
		return
	}
	JSON(w, http.StatusOK, result)
}

// GET /api/transactions/{id}
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	txnID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	hhID := middleware.HouseholdIDFromCtx(r.Context())
	txn, err := h.txnSvc.Get(r.Context(), txnID, hhID)
	if err != nil {
		ErrorJSON(w, http.StatusNotFound, "transaction not found")
		return
	}
	JSON(w, http.StatusOK, txn)
}

// PUT /api/transactions/{id}
func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	txnID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	var req model.UpdateTransactionRequest
	if err := Decode(r, &req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := middleware.UserIDFromCtx(r.Context())
	hhID := middleware.HouseholdIDFromCtx(r.Context())

	txn, err := h.txnSvc.Update(r.Context(), txnID, hhID, userID, req)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	JSON(w, http.StatusOK, txn)
}

// DELETE /api/transactions/{id}
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	txnID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	hhID := middleware.HouseholdIDFromCtx(r.Context())
	if err := h.txnSvc.Delete(r.Context(), txnID, hhID); err != nil {
		ErrorJSON(w, http.StatusNotFound, "transaction not found")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"message": "transaction deleted"})
}
