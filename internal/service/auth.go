package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/howallet/howallet/internal/config"
	"github.com/howallet/howallet/internal/model"
	"github.com/howallet/howallet/internal/repository/postgres"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type AuthService struct {
	repos *postgres.Repos
	jwt   *config.JWTConfig
}

func NewAuthService(repos *postgres.Repos, jwtCfg *config.JWTConfig) *AuthService {
	return &AuthService{repos: repos, jwt: jwtCfg}
}

// Register creates a new user, a default household, and returns tokens.
func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, error) {
	// Check if email is taken
	_, err := s.repos.Users.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrEmailTaken
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check email: %w", err)
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var user model.User
	err = s.repos.RunInTx(ctx, func(txCtx context.Context) error {
		txRepos := postgres.TxReposFromCtx(txCtx)

		var txErr error
		user, txErr = txRepos.Users.Create(txCtx, req.Email, string(hash), req.Name)
		if txErr != nil {
			return fmt.Errorf("create user: %w", txErr)
		}

		hh, txErr := txRepos.Households.Create(txCtx, req.Name+"'s Wallet", user.ID)
		if txErr != nil {
			return fmt.Errorf("create household: %w", txErr)
		}

		txErr = txRepos.Households.AddMember(txCtx, hh.ID, user.ID, model.HouseholdRoleOwner)
		if txErr != nil {
			return fmt.Errorf("add household member: %w", txErr)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateAndStoreRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// Login authenticates a user and returns tokens.
func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	user, err := s.repos.Users.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateAndStoreRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// Refresh validates a refresh token and issues a new access + refresh pair.
func (s *AuthService) Refresh(ctx context.Context, rawToken string) (*model.AuthResponse, error) {
	h := hashToken(rawToken)

	rt, err := s.repos.RefreshTokens.GetByHash(ctx, h)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if rt.ExpiresAt.Before(time.Now()) {
		_ = s.repos.RefreshTokens.Delete(ctx, h)
		return nil, ErrInvalidToken
	}

	// Delete the old refresh token (rotation)
	_ = s.repos.RefreshTokens.Delete(ctx, h)

	user, err := s.repos.Users.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	newRefresh, err := s.generateAndStoreRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefresh,
		User:         user,
	}, nil
}

// Logout deletes all refresh tokens for the user.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.repos.RefreshTokens.DeleteByUser(ctx, userID)
}

// --- token helpers ---

func (s *AuthService) generateAccessToken(userID uuid.UUID, email string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"iat":   now.Unix(),
		"exp":   now.Add(s.jwt.AccessTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwt.Secret))
}

func (s *AuthService) generateAndStoreRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	raw := generateRandomToken(32)
	h := hashToken(raw)

	err := s.repos.RefreshTokens.Create(ctx, userID, h, time.Now().Add(s.jwt.RefreshTTL))
	if err != nil {
		return "", fmt.Errorf("store refresh token: %w", err)
	}

	return raw, nil
}

func generateRandomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand.Read failed: %v", err))
	}
	return hex.EncodeToString(b)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
