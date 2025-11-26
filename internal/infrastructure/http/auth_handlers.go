package http

import (
	"encoding/json"
	"net/http"

	"apocapoc-api/internal/application/commands"
	"apocapoc-api/internal/application/queries"
	"apocapoc-api/internal/infrastructure/auth"
	"apocapoc-api/internal/shared/errors"
)

type AuthHandlers struct {
	registerHandler *commands.RegisterUserHandler
	loginHandler    *queries.LoginUserHandler
	jwtService      *auth.JWTService
}

func NewAuthHandlers(
	registerHandler *commands.RegisterUserHandler,
	loginHandler *queries.LoginUserHandler,
	jwtService *auth.JWTService,
) *AuthHandlers {
	return &AuthHandlers{
		registerHandler: registerHandler,
		loginHandler:    loginHandler,
		jwtService:      jwtService,
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
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration data"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
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

	respondJSON(w, http.StatusCreated, AuthResponse{
		Token:  token,
		UserID: userID,
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
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

	respondJSON(w, http.StatusOK, AuthResponse{
		Token:  token,
		UserID: result.UserID,
	})
}
