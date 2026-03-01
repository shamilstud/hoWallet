package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/howallet/howallet/internal/middleware"
	"github.com/howallet/howallet/internal/model"
	"github.com/howallet/howallet/internal/service"
)

type AccountHandler struct {
	accSvc *service.AccountService
}

func NewAccountHandler(accSvc *service.AccountService) *AccountHandler {
	return &AccountHandler{accSvc: accSvc}
}

// POST /api/accounts
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAccountRequest
	if err := Decode(r, &req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		ErrorJSON(w, http.StatusBadRequest, "name is required")
		return
	}

	userID := middleware.UserIDFromCtx(r.Context())
	hhID := middleware.HouseholdIDFromCtx(r.Context())

	acc, err := h.accSvc.Create(r.Context(), hhID, userID, req)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to create account")
		return
	}
	JSON(w, http.StatusCreated, acc)
}

// GET /api/accounts
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	hhID := middleware.HouseholdIDFromCtx(r.Context())

	accounts, err := h.accSvc.List(r.Context(), hhID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}
	JSON(w, http.StatusOK, accounts)
}

// GET /api/accounts/{id}
func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	accID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid account id")
		return
	}

	hhID := middleware.HouseholdIDFromCtx(r.Context())
	acc, err := h.accSvc.Get(r.Context(), accID, hhID)
	if err != nil {
		ErrorJSON(w, http.StatusNotFound, "account not found")
		return
	}
	JSON(w, http.StatusOK, acc)
}

// PUT /api/accounts/{id}
func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	accID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid account id")
		return
	}

	var req model.UpdateAccountRequest
	if err := Decode(r, &req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	hhID := middleware.HouseholdIDFromCtx(r.Context())
	acc, err := h.accSvc.Update(r.Context(), accID, hhID, req)
	if err != nil {
		ErrorJSON(w, http.StatusNotFound, "account not found")
		return
	}
	JSON(w, http.StatusOK, acc)
}

// DELETE /api/accounts/{id}
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	accID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid account id")
		return
	}

	hhID := middleware.HouseholdIDFromCtx(r.Context())
	err = h.accSvc.Delete(r.Context(), accID, hhID)
	if err != nil {
		if errors.Is(err, service.ErrAccountHasTransactions) {
			ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}
		ErrorJSON(w, http.StatusInternalServerError, "failed to delete account")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"message": "account deleted"})
}
