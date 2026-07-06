package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/lam-thinh/customer-oil-change-management/internal/auth"
	"github.com/lam-thinh/customer-oil-change-management/internal/dto"
	"github.com/lam-thinh/customer-oil-change-management/internal/service"
)

// AuthServicer is the interface the AuthHandler depends on.
// Keeping this narrow makes the handler trivially testable.
type AuthServicer interface {
	Register(ctx context.Context, email, password string) (*service.RegisterResult, error)
	Login(ctx context.Context, email, password string) (*service.LoginResult, error)
	Refresh(ctx context.Context, refreshToken string) (*service.RefreshResult, error)
	Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error
}

// AuthHandler holds the dependencies for authentication HTTP handlers.
type AuthHandler struct {
	svc AuthServicer
	log *slog.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc AuthServicer, log *slog.Logger) *AuthHandler {
	return &AuthHandler{
		svc: svc,
		log: log,
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func mapServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrEmailTaken):
		Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		Error(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrSessionNotFound),
		errors.Is(err, service.ErrSessionRevoked),
		errors.Is(err, service.ErrSessionExpired):
		Error(w, http.StatusUnauthorized, err.Error())
	default:
		Error(w, http.StatusInternalServerError, "internal server error")
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// Register godoc
//
//	@Summary      Register a new user
//	@Description  Create a new account with email and password
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        request  body      dto.RegisterRequest   true  "Registration credentials"
//	@Success      201      {object}  dto.RegisterResponse  "User created"
//	@Failure      400      {object}  dto.ErrorResponse     "Malformed JSON"
//	@Failure      409      {object}  dto.ErrorResponse     "Email already registered"
//	@Failure      422      {object}  dto.ErrorResponse     "Validation failed"
//	@Router       /v1/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if !decodeAndValidate(w, r, &req) {
		return
	}

	result, err := h.svc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		h.log.ErrorContext(r.Context(), "register failed", "error", err)
		mapServiceError(w, err)
		return
	}

	JSON(w, http.StatusCreated, dto.RegisterResponse{
		ID:    result.User.ID.String(),
		Email: result.User.Email,
	})
}

// Login godoc
//
//	@Summary      Login
//	@Description  Authenticate with email and password; returns an access token and a refresh token
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        request  body      dto.LoginRequest   true  "Login credentials"
//	@Success      200      {object}  dto.LoginResponse  "Tokens returned"
//	@Failure      400      {object}  dto.ErrorResponse  "Malformed JSON"
//	@Failure      401      {object}  dto.ErrorResponse  "Invalid credentials"
//	@Failure      422      {object}  dto.ErrorResponse  "Validation failed"
//	@Router       /v1/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if !decodeAndValidate(w, r, &req) {
		return
	}

	result, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.log.ErrorContext(r.Context(), "login failed", "error", err)
		mapServiceError(w, err)
		return
	}

	JSON(w, http.StatusOK, dto.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		UserID:       result.User.ID.String(),
	})
}

// Refresh godoc
//
//	@Summary      Refresh tokens
//	@Description  Exchange a valid refresh token for a new access token (old refresh token is revoked)
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        request  body      dto.RefreshRequest   true  "Refresh token"
//	@Success      200      {object}  dto.RefreshResponse  "New tokens returned"
//	@Failure      400      {object}  dto.ErrorResponse    "Malformed JSON"
//	@Failure      401      {object}  dto.ErrorResponse    "Token not found, revoked, or expired"
//	@Failure      422      {object}  dto.ErrorResponse    "Validation failed"
//	@Router       /v1/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if !decodeAndValidate(w, r, &req) {
		return
	}

	result, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		h.log.ErrorContext(r.Context(), "refresh failed", "error", err)
		mapServiceError(w, err)
		return
	}

	JSON(w, http.StatusOK, dto.RefreshResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

// Logout godoc
//
//	@Summary      Logout
//	@Description  Revoke a refresh token. Requires a valid JWT access token in the Authorization header.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        request  body  dto.LogoutRequest  true  "Refresh token to revoke"
//	@Success      204      "Logged out"
//	@Failure      400      {object}  dto.ErrorResponse  "Malformed JSON"
//	@Failure      401      {object}  dto.ErrorResponse  "Not authenticated or token not found"
//	@Failure      422      {object}  dto.ErrorResponse  "Validation failed"
//	@Router       /v1/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.LogoutRequest
	if !decodeAndValidate(w, r, &req) {
		return
	}

	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	if err := h.svc.Logout(r.Context(), claims.UserID, req.RefreshToken); err != nil {
		h.log.ErrorContext(r.Context(), "logout failed", "error", err)
		mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
