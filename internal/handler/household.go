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

type HouseholdHandler struct {
	hhSvc *service.HouseholdService
}

func NewHouseholdHandler(hhSvc *service.HouseholdService) *HouseholdHandler {
	return &HouseholdHandler{hhSvc: hhSvc}
}

// POST /api/households
func (h *HouseholdHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateHouseholdRequest
	if err := Decode(r, &req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		ErrorJSON(w, http.StatusBadRequest, "name is required")
		return
	}

	userID := middleware.UserIDFromCtx(r.Context())
	hh, err := h.hhSvc.Create(r.Context(), userID, req)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to create household")
		return
	}
	JSON(w, http.StatusCreated, hh)
}

// GET /api/households
func (h *HouseholdHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	list, err := h.hhSvc.List(r.Context(), userID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to list households")
		return
	}
	JSON(w, http.StatusOK, list)
}

// GET /api/households/{id}/members
func (h *HouseholdHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	hhID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid household id")
		return
	}

	userID := middleware.UserIDFromCtx(r.Context())
	if err := h.hhSvc.CheckMembership(r.Context(), hhID, userID); err != nil {
		ErrorJSON(w, http.StatusForbidden, "not a member")
		return
	}

	members, err := h.hhSvc.ListMembers(r.Context(), hhID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to list members")
		return
	}
	JSON(w, http.StatusOK, members)
}

// POST /api/households/{id}/invite
func (h *HouseholdHandler) Invite(w http.ResponseWriter, r *http.Request) {
	hhID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid household id")
		return
	}

	var req model.InviteRequest
	if err := Decode(r, &req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" {
		ErrorJSON(w, http.StatusBadRequest, "email is required")
		return
	}

	userID := middleware.UserIDFromCtx(r.Context())
	inv, err := h.hhSvc.Invite(r.Context(), hhID, userID, req.Email)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotHouseholdOwner):
			ErrorJSON(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrAlreadyMember):
			ErrorJSON(w, http.StatusConflict, err.Error())
		default:
			ErrorJSON(w, http.StatusInternalServerError, "failed to send invitation")
		}
		return
	}
	JSON(w, http.StatusCreated, inv)
}

// POST /api/invitations/{token}/accept
func (h *HouseholdHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		ErrorJSON(w, http.StatusBadRequest, "token is required")
		return
	}

	userID := middleware.UserIDFromCtx(r.Context())
	if err := h.hhSvc.AcceptInvitation(r.Context(), token, userID); err != nil {
		if errors.Is(err, service.ErrInvitationInvalid) {
			ErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		ErrorJSON(w, http.StatusInternalServerError, "failed to accept invitation")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"message": "invitation accepted"})
}

// DELETE /api/households/{id}/members/{userId}
func (h *HouseholdHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	hhID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid household id")
		return
	}
	targetUID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}

	ownerID := middleware.UserIDFromCtx(r.Context())
	if err := h.hhSvc.RemoveMember(r.Context(), hhID, ownerID, targetUID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotHouseholdOwner):
			ErrorJSON(w, http.StatusForbidden, err.Error())
		default:
			ErrorJSON(w, http.StatusInternalServerError, "failed to remove member")
		}
		return
	}
	JSON(w, http.StatusOK, map[string]string{"message": "member removed"})
}
