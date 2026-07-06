package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/lam-thinh/customer-oil-change-management/internal/auth"
	db "github.com/lam-thinh/customer-oil-change-management/internal/db/sqlc"
)

// Sentinel errors returned by AuthService methods.
var (
	ErrEmailTaken       = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrSessionNotFound  = errors.New("session not found")
	ErrSessionRevoked   = errors.New("session has been revoked")
	ErrSessionExpired   = errors.New("session has expired")
)

// AuthRepository is the subset of db.Querier used by AuthService.
// Using an interface here keeps the service testable without a real database.
type AuthRepository interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error)
	CreateSession(ctx context.Context, arg db.CreateSessionParams) (db.Session, error)
	GetSessionByToken(ctx context.Context, refreshToken string) (db.Session, error)
	UpdateSessionRevoked(ctx context.Context, refreshToken string) error
	DeleteAllSessionsForUser(ctx context.Context, userID uuid.UUID) error
}

// AuthService contains all business logic related to authentication.
type AuthService struct {
	repo               AuthRepository
	jwtSecret          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	log                *slog.Logger
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	repo AuthRepository,
	jwtSecret string,
	accessTokenExpiry time.Duration,
	refreshTokenExpiry time.Duration,
	log *slog.Logger,
) *AuthService {
	return &AuthService{
		repo:               repo,
		jwtSecret:          jwtSecret,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
		log:                log,
	}
}

// RegisterResult is the data returned after a successful registration.
type RegisterResult struct {
	User db.User
}

// Register validates uniqueness, hashes the password, and stores the new user.
func (s *AuthService) Register(ctx context.Context, email, password string) (*RegisterResult, error) {
	// Check for existing email.
	_, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, ErrEmailTaken
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("service: check email: %w", err)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("service: hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
	})
	if err != nil {
		return nil, fmt.Errorf("service: create user: %w", err)
	}

	s.log.InfoContext(ctx, "user registered", "user_id", user.ID)
	return &RegisterResult{User: user}, nil
}

// LoginResult is the data returned after a successful login.
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	User         db.User
}

// Login verifies credentials, generates tokens, and stores the refresh token.
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("service: get user: %w", err)
	}

	if !auth.CheckPassword(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, s.jwtSecret, s.accessTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("service: generate access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("service: generate refresh token: %w", err)
	}

	_, err = s.repo.CreateSession(ctx, db.CreateSessionParams{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.refreshTokenExpiry),
	})
	if err != nil {
		return nil, fmt.Errorf("service: create session: %w", err)
	}

	s.log.InfoContext(ctx, "user logged in", "user_id", user.ID)
	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// RefreshResult is the data returned after a successful token refresh.
type RefreshResult struct {
	AccessToken  string
	RefreshToken string // new refresh token (rotation)
}

// Refresh validates the refresh token and issues a new access token.
// It also rotates the refresh token (issues a new one, revokes the old one).
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	session, err := s.repo.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("service: get session: %w", err)
	}

	if session.IsRevoked {
		return nil, ErrSessionRevoked
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	// Revoke the old refresh token (rotation).
	if err := s.repo.UpdateSessionRevoked(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("service: revoke old session: %w", err)
	}

	newAccessToken, err := auth.GenerateAccessToken(session.UserID, s.jwtSecret, s.accessTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("service: generate access token: %w", err)
	}

	newRefreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("service: generate refresh token: %w", err)
	}

	_, err = s.repo.CreateSession(ctx, db.CreateSessionParams{
		UserID:       session.UserID,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(s.refreshTokenExpiry),
	})
	if err != nil {
		return nil, fmt.Errorf("service: create new session: %w", err)
	}

	s.log.InfoContext(ctx, "tokens refreshed", "user_id", session.UserID)
	return &RefreshResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout revokes the given refresh token for the authenticated user.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	session, err := s.repo.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSessionNotFound
		}
		return fmt.Errorf("service: get session: %w", err)
	}

	// Ensure the token belongs to the requesting user.
	if session.UserID != userID {
		return ErrSessionNotFound
	}

	if err := s.repo.UpdateSessionRevoked(ctx, refreshToken); err != nil {
		return fmt.Errorf("service: revoke session: %w", err)
	}

	s.log.InfoContext(ctx, "user logged out", "user_id", userID)
	return nil
}
