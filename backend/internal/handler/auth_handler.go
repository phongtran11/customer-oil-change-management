package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

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
	svc          AuthServicer
	log          *slog.Logger
	secureCookie bool
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc AuthServicer, log *slog.Logger, secureCookie bool) *AuthHandler {
	return &AuthHandler{
		svc:          svc,
		log:          log,
		secureCookie: secureCookie,
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
//	@Description  Authenticate with email and password; returns an access token and sets a secure cookie containing a refresh token
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        request  body      dto.LoginRequest   true  "Login credentials"
//	@Success      200      {object}  dto.LoginResponse  "Access token returned and cookie set"
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

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		Path:     "/",
		Expires:  time.Now().Add(result.RefreshTokenExpiry),
		MaxAge:   int(result.RefreshTokenExpiry.Seconds()),
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	JSON(w, http.StatusOK, dto.LoginResponse{
		AccessToken: result.AccessToken,
		UserID:      result.User.ID.String(),
	})
}

// Refresh godoc
//
//	@Summary      Refresh tokens
//	@Description  Exchange a valid refresh token cookie for a new access token (old refresh token is revoked)
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Success      200      {object}  dto.RefreshResponse  "New access token returned and cookie set"
//	@Failure      401      {object}  dto.ErrorResponse    "Cookie not found, token revoked, or expired"
//	@Router       /v1/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		Error(w, http.StatusUnauthorized, "missing refresh token")
		return
	}
	rawRefreshToken := cookie.Value

	result, err := h.svc.Refresh(r.Context(), rawRefreshToken)
	if err != nil {
		h.log.ErrorContext(r.Context(), "refresh failed", "error", err)
		mapServiceError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		Path:     "/",
		Expires:  time.Now().Add(result.RefreshTokenExpiry),
		MaxAge:   int(result.RefreshTokenExpiry.Seconds()),
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	JSON(w, http.StatusOK, dto.RefreshResponse{
		AccessToken: result.AccessToken,
	})
}

// Logout godoc
//
//	@Summary      Logout
//	@Description  Revoke a refresh token from cookie. Requires a valid JWT access token in the Authorization header.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Success      204      "Logged out and cookie cleared"
//	@Failure      401      {object}  dto.ErrorResponse  "Not authenticated or token not found"
//	@Router       /v1/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		Error(w, http.StatusUnauthorized, "missing refresh token")
		return
	}
	rawRefreshToken := cookie.Value

	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	if err := h.svc.Logout(r.Context(), claims.UserID, rawRefreshToken); err != nil {
		h.log.ErrorContext(r.Context(), "logout failed", "error", err)
		mapServiceError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}
