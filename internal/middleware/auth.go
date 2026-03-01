package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/howallet/howallet/internal/config"
)

type contextKey string

const (
	ContextKeyUserID      contextKey = "user_id"
	ContextKeyHouseholdID contextKey = "household_id"
)

// UserIDFromCtx extracts the authenticated user ID from context.
func UserIDFromCtx(ctx context.Context) uuid.UUID {
	if v, ok := ctx.Value(ContextKeyUserID).(uuid.UUID); ok {
		return v
	}
	return uuid.Nil
}

// HouseholdIDFromCtx extracts the active household ID from context.
func HouseholdIDFromCtx(ctx context.Context) uuid.UUID {
	if v, ok := ctx.Value(ContextKeyHouseholdID).(uuid.UUID); ok {
		return v
	}
	return uuid.Nil
}

// JWTAuth validates the Bearer token from the Authorization header.
func JWTAuth(cfg *config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(cfg.Secret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"invalid token claims"}`, http.StatusUnauthorized)
				return
			}

			userIDStr, _ := claims["sub"].(string)
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				http.Error(w, `{"error":"invalid user id in token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MembershipChecker is a function that verifies a user belongs to a household.
type MembershipChecker func(ctx context.Context, householdID, userID uuid.UUID) error

// HouseholdCtx reads X-Household-ID header, puts it into context,
// and verifies the authenticated user is a member of that household.
func HouseholdCtx(checkMembership MembershipChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hhIDStr := r.Header.Get("X-Household-ID")
			if hhIDStr == "" {
				http.Error(w, `{"error":"missing X-Household-ID header"}`, http.StatusBadRequest)
				return
			}

			hhID, err := uuid.Parse(hhIDStr)
			if err != nil {
				http.Error(w, `{"error":"invalid X-Household-ID"}`, http.StatusBadRequest)
				return
			}

			userID := UserIDFromCtx(r.Context())
			if err := checkMembership(r.Context(), hhID, userID); err != nil {
				http.Error(w, `{"error":"not a member of this household"}`, http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyHouseholdID, hhID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
