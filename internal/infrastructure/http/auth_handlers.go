package http

import (
	"encoding/json"
	"net/http"
	"time"

	"apocapoc-api/internal/application/commands"
	"apocapoc-api/internal/application/queries"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/infrastructure/auth"
	"apocapoc-api/internal/shared/errors"
)

type AuthHandlers struct {
	registerHandler        *commands.RegisterUserHandler
	loginHandler           *queries.LoginUserHandler
	refreshTokenHandler    *queries.RefreshTokenHandler
	revokeTokenHandler     *commands.RevokeTokenHandler
	revokeAllTokensHandler *commands.RevokeAllTokensHandler
	jwtService             *auth.JWTService
	refreshTokenRepo       repositories.RefreshTokenRepository
	refreshTokenExpiry     time.Duration
}

func NewAuthHandlers(
	registerHandler *commands.RegisterUserHandler,
	loginHandler *queries.LoginUserHandler,
	refreshTokenHandler *queries.RefreshTokenHandler,
	revokeTokenHandler *commands.RevokeTokenHandler,
	revokeAllTokensHandler *commands.RevokeAllTokensHandler,
	jwtService *auth.JWTService,
	refreshTokenRepo repositories.RefreshTokenRepository,
	refreshTokenExpiry time.Duration,
) *AuthHandlers {
	return &AuthHandlers{
		registerHandler:        registerHandler,
		loginHandler:           loginHandler,
		refreshTokenHandler:    refreshTokenHandler,
		revokeTokenHandler:     revokeTokenHandler,
		revokeAllTokensHandler: revokeAllTokensHandler,
		jwtService:             jwtService,
		refreshTokenRepo:       refreshTokenRepo,
		refreshTokenExpiry:     refreshTokenExpiry,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Timezone string `json:"timezone"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account and receive both access token and refresh token. Store both tokens securely - the refresh token is used to obtain new access tokens when they expire.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration data (password requires: min 8 chars, uppercase, lowercase, digit, special char)"
// @Success 201 {object} AuthResponse "Returns access token, refresh token, and user ID"
// @Failure 400 {object} ErrorResponse "Invalid input: email format, password requirements, or timezone"
// @Failure 409 {object} ErrorResponse "Email already registered"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := commands.RegisterUserCommand{
		Email:    req.Email,
		Password: req.Password,
		Timezone: req.Timezone,
	}

	userID, err := h.registerHandler.Handle(r.Context(), cmd)
	if err != nil {
		if err == errors.ErrInvalidInput {
			respondError(w, http.StatusBadRequest, "Invalid email or password (min 8 characters)")
			return
		}
		if err == errors.ErrAlreadyExists {
			respondError(w, http.StatusConflict, "Email already registered")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	token, err := h.jwtService.GenerateToken(userID, req.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	refreshToken, err := queries.CreateRefreshToken(userID, h.refreshTokenExpiry)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create refresh token")
		return
	}

	if err := h.refreshTokenRepo.Create(r.Context(), refreshToken); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save refresh token")
		return
	}

	respondJSON(w, http.StatusCreated, AuthResponse{
		Token:        token,
		RefreshToken: refreshToken.Token,
		UserID:       userID,
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password. Returns both access token and refresh token. The access token is used for API requests, the refresh token is used to obtain new access tokens when they expire.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse "Returns access token, refresh token, and user ID"
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Invalid email or password"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	query := queries.LoginUserQuery{
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := h.loginHandler.Handle(r.Context(), query)
	if err != nil {
		if err == errors.ErrNotFound || err == errors.ErrInvalidInput {
			respondError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to login")
		return
	}

	token, err := h.jwtService.GenerateToken(result.UserID, result.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	refreshToken, err := queries.CreateRefreshToken(result.UserID, h.refreshTokenExpiry)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create refresh token")
		return
	}

	if err := h.refreshTokenRepo.Create(r.Context(), refreshToken); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save refresh token")
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		Token:        token,
		RefreshToken: refreshToken.Token,
		UserID:       result.UserID,
	})
}

// Refresh godoc
// @Summary Refresh access token
// @Description Exchange a valid refresh token for a new access token and refresh token pair. IMPORTANT: The old refresh token is automatically revoked and you receive a NEW refresh token - always update both tokens in storage. Use this endpoint when the access token expires to maintain the user session without requiring re-login.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Current refresh token"
// @Success 200 {object} AuthResponse "Returns NEW access token and NEW refresh token - the old refresh token is now invalid"
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Invalid or expired refresh token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	query := queries.RefreshTokenQuery{
		RefreshToken: req.RefreshToken,
	}

	result, err := h.refreshTokenHandler.Handle(r.Context(), query)
	if err != nil {
		if err == errors.ErrNotFound || err == errors.ErrInvalidInput {
			respondError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to refresh token")
		return
	}

	token, err := h.jwtService.GenerateToken(result.UserID, result.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	newRefreshToken, err := queries.CreateRefreshToken(result.UserID, h.refreshTokenExpiry)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create refresh token")
		return
	}

	if err := h.refreshTokenRepo.Create(r.Context(), newRefreshToken); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save refresh token")
		return
	}

	if err := h.refreshTokenRepo.RevokeByToken(r.Context(), req.RefreshToken); err != nil {
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		Token:        token,
		RefreshToken: newRefreshToken.Token,
		UserID:       result.UserID,
	})
}

// Logout godoc
// @Summary Logout user
// @Description Revoke the refresh token to invalidate the user session. After logout, the refresh token cannot be used to obtain new access tokens. The user will need to login again. Always call this endpoint before clearing tokens from client storage to ensure proper session termination.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Refresh token to revoke"
// @Success 200 {object} map[string]string "Successfully logged out - the refresh token is now invalid"
// @Failure 400 {object} ErrorResponse "Invalid request body or missing refresh token"
// @Failure 404 {object} ErrorResponse "Refresh token not found (already revoked or never existed)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/logout [post]
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := commands.RevokeTokenCommand{
		RefreshToken: req.RefreshToken,
	}

	err := h.revokeTokenHandler.Handle(r.Context(), cmd)
	if err != nil {
		if err == errors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Refresh token not found")
			return
		}
		if err == errors.ErrInvalidInput {
			respondError(w, http.StatusBadRequest, "Invalid refresh token")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to revoke token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Successfully logged out",
	})
}
