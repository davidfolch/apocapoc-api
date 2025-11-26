package http

import (
	"encoding/json"
	"net/http"

	"habit-tracker-api/internal/application/commands"
	"habit-tracker-api/internal/application/queries"
	"habit-tracker-api/internal/infrastructure/auth"
	"habit-tracker-api/internal/shared/errors"
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
