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
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/howallet/howallet/internal/config"
	db "github.com/howallet/howallet/internal/db"
	"github.com/howallet/howallet/internal/model"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type AuthService struct {
	queries *db.Queries
	pool    *pgxpool.Pool
	jwt     *config.JWTConfig
}

func NewAuthService(pool *pgxpool.Pool, queries *db.Queries, jwtCfg *config.JWTConfig) *AuthService {
	return &AuthService{queries: queries, pool: pool, jwt: jwtCfg}
}

// Register creates a new user, a default household, and returns tokens.
func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, error) {
	// Check if email is taken
	_, err := s.queries.GetUserByEmail(ctx, req.Email)
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

	// Use a DB transaction for atomicity
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	// Create user
	dbUser, err := qtx.CreateUser(ctx, db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Create default household
	hh, err := qtx.CreateHousehold(ctx, db.CreateHouseholdParams{
		Name:    req.Name + "'s Wallet",
		OwnerID: dbUser.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("create household: %w", err)
	}

	// Add user as owner
	err = qtx.AddHouseholdMember(ctx, db.AddHouseholdMemberParams{
		HouseholdID: hh.ID,
		UserID:      dbUser.ID,
		Role:        db.HouseholdRoleOwner,
	})
	if err != nil {
		return nil, fmt.Errorf("add household member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(dbUser.ID, dbUser.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateAndStoreRefreshToken(ctx, dbUser.ID)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserModel(dbUser),
	}, nil
}

// Login authenticates a user and returns tokens.
func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	dbUser, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.generateAccessToken(dbUser.ID, dbUser.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateAndStoreRefreshToken(ctx, dbUser.ID)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserModel(dbUser),
	}, nil
}

// Refresh validates a refresh token and issues a new access + refresh pair.
func (s *AuthService) Refresh(ctx context.Context, rawToken string) (*model.AuthResponse, error) {
	hash := hashToken(rawToken)

	rt, err := s.queries.GetRefreshToken(ctx, hash)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if rt.ExpiresAt.Time.Before(time.Now()) {
		_ = s.queries.DeleteRefreshToken(ctx, hash)
		return nil, ErrInvalidToken
	}

	// Delete the old refresh token (rotation)
	_ = s.queries.DeleteRefreshToken(ctx, hash)

	user, err := s.queries.GetUserByID(ctx, rt.UserID)
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
		User:         toUserModel(user),
	}, nil
}

// Logout deletes all refresh tokens for the user.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.queries.DeleteUserRefreshTokens(ctx, userID)
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
	hash := hashToken(raw)

	err := s.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(s.jwt.RefreshTTL),
	})
	if err != nil {
		return "", fmt.Errorf("store refresh token: %w", err)
	}

	return raw, nil
}

func generateRandomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func toUserModel(u db.User) model.User {
	return model.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}
