package handler

import (
	"errors"
	"net/http"

	"github.com/howallet/howallet/internal/middleware"
	"github.com/howallet/howallet/internal/model"
	"github.com/howallet/howallet/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := Decode(r, &req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		ErrorJSON(w, http.StatusBadRequest, "email, password and name are required")
		return
	}

	if len(req.Password) < 8 {
		ErrorJSON(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	resp, err := h.authSvc.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}
		ErrorJSON(w, http.StatusInternalServerError, "registration failed")
		return
	}

	JSON(w, http.StatusCreated, resp)
}

// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := Decode(r, &req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		ErrorJSON(w, http.StatusBadRequest, "email and password are required")
		return
	}

	resp, err := h.authSvc.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			ErrorJSON(w, http.StatusUnauthorized, err.Error())
			return
		}
		ErrorJSON(w, http.StatusInternalServerError, "login failed")
		return
	}

	JSON(w, http.StatusOK, resp)
}

// POST /auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshRequest
	if err := Decode(r, &req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		ErrorJSON(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	resp, err := h.authSvc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			ErrorJSON(w, http.StatusUnauthorized, err.Error())
			return
		}
		ErrorJSON(w, http.StatusInternalServerError, "refresh failed")
		return
	}

	JSON(w, http.StatusOK, resp)
}

// POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	if err := h.authSvc.Logout(r.Context(), userID); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "logout failed")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}
