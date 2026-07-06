package dto

// ── Auth Request DTOs ────────────────────────────────────────────────────────

// RegisterRequest is the request body for POST /register.
type RegisterRequest struct {
	Email    string `json:"email"    validate:"required,email,max=255" example:"alice@example.com"`
	Password string `json:"password" validate:"required,min=8,max=128"  example:"securePass123"`
}

// LoginRequest is the request body for POST /login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email" example:"alice@example.com"`
	Password string `json:"password" validate:"required"       example:"securePass123"`
}

// RefreshRequest is the request body for POST /refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"a3f7c2e1d4b9..."`
}

// LogoutRequest is the request body for POST /logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"a3f7c2e1d4b9..."`
}

// ── Auth Response DTOs ───────────────────────────────────────────────────────

// RegisterResponse is the response body for POST /register.
type RegisterResponse struct {
	ID    string `json:"id"    example:"550e8400-e29b-41d4-a716-446655440000"`
	Email string `json:"email" example:"alice@example.com"`
}

// LoginResponse is the response body for POST /login.
type LoginResponse struct {
	AccessToken  string `json:"access_token"  example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"a3f7c2e1d4b9..."`
	UserID       string `json:"user_id"       example:"550e8400-e29b-41d4-a716-446655440000"`
}

// RefreshResponse is the response body for POST /refresh.
type RefreshResponse struct {
	AccessToken  string `json:"access_token"  example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"b8e1a4f2c6d3..."`
}

// ── Common Response DTOs ─────────────────────────────────────────────────────

// ErrorResponse is the standard error envelope returned on any failure.
type ErrorResponse struct {
	Error string `json:"error" example:"invalid email or password"`
}
