package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"apocapoc-api/internal/application/commands"
	"apocapoc-api/internal/application/queries"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/infrastructure/auth"
	appErrors "apocapoc-api/internal/shared/errors"
)

type AuthHandlers struct {
	registerHandler                *commands.RegisterUserHandler
	loginHandler                   *queries.LoginUserHandler
	refreshTokenHandler            *queries.RefreshTokenHandler
	revokeTokenHandler             *commands.RevokeTokenHandler
	revokeAllTokensHandler         *commands.RevokeAllTokensHandler
	verifyEmailHandler             *commands.VerifyEmailHandler
	resendVerificationEmailHandler *commands.ResendVerificationEmailHandler
	requestPasswordResetHandler    *commands.RequestPasswordResetHandler
	resetPasswordHandler           *commands.ResetPasswordHandler
	jwtService                     *auth.JWTService
	refreshTokenRepo               repositories.RefreshTokenRepository
	refreshTokenExpiry             time.Duration
}

func NewAuthHandlers(
	registerHandler *commands.RegisterUserHandler,
	loginHandler *queries.LoginUserHandler,
	refreshTokenHandler *queries.RefreshTokenHandler,
	revokeTokenHandler *commands.RevokeTokenHandler,
	revokeAllTokensHandler *commands.RevokeAllTokensHandler,
	verifyEmailHandler *commands.VerifyEmailHandler,
	resendVerificationEmailHandler *commands.ResendVerificationEmailHandler,
	requestPasswordResetHandler *commands.RequestPasswordResetHandler,
	resetPasswordHandler *commands.ResetPasswordHandler,
	jwtService *auth.JWTService,
	refreshTokenRepo repositories.RefreshTokenRepository,
	refreshTokenExpiry time.Duration,
) *AuthHandlers {
	return &AuthHandlers{
		registerHandler:                registerHandler,
		loginHandler:                   loginHandler,
		refreshTokenHandler:            refreshTokenHandler,
		revokeTokenHandler:             revokeTokenHandler,
		revokeAllTokensHandler:         revokeAllTokensHandler,
		verifyEmailHandler:             verifyEmailHandler,
		resendVerificationEmailHandler: resendVerificationEmailHandler,
		requestPasswordResetHandler:    requestPasswordResetHandler,
		resetPasswordHandler:           resetPasswordHandler,
		jwtService:                     jwtService,
		refreshTokenRepo:               refreshTokenRepo,
		refreshTokenExpiry:             refreshTokenExpiry,
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

type RegisterResponse struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account. If email verification is enabled, you will receive a verification email. Otherwise, you can login immediately.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration data (password requires: min 8 chars, uppercase, lowercase, digit, special char)"
// @Success 201 {object} RegisterResponse "Returns user ID and message about next steps"
// @Failure 400 {object} ErrorResponse "Invalid input: email format, password requirements, or timezone"
// @Failure 403 {object} ErrorResponse "Registration is closed"
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

	result, err := h.registerHandler.Handle(r.Context(), cmd)
	if err != nil {
		if errors.Is(err, appErrors.ErrInvalidInput) {
			respondValidationError(w, err)
			return
		}
		if err == appErrors.ErrAlreadyExists {
			respondError(w, http.StatusConflict, "Email already registered")
			return
		}
		if err == appErrors.ErrRegistrationClosed {
			respondError(w, http.StatusForbidden, "Registration is currently closed")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	var message string
	if result.EmailVerificationRequired {
		message = "Registration successful. Please check your email to verify your account."
	} else {
		message = "Registration successful. You can now login."
	}

	respondJSON(w, http.StatusCreated, RegisterResponse{
		UserID:  result.UserID,
		Message: message,
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
// @Failure 403 {object} ErrorResponse "Email not verified"
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
		if err == appErrors.ErrNotFound || err == appErrors.ErrInvalidInput {
			respondError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		if err == appErrors.ErrEmailNotVerified {
			respondError(w, http.StatusForbidden, "Please verify your email before logging in")
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
		if err == appErrors.ErrNotFound || err == appErrors.ErrInvalidInput {
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
		if err == appErrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Refresh token not found")
			return
		}
		if err == appErrors.ErrInvalidInput {
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

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

type ResendVerificationRequest struct {
	Email string `json:"email"`
}

// VerifyEmail godoc
// @Summary Verify email address
// @Description Verify user email address using the token sent via email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "Verification token"
// @Success 200 {object} map[string]string "Email verified successfully"
// @Failure 400 {object} ErrorResponse "Invalid or expired token"
// @Failure 409 {object} ErrorResponse "Email already verified"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/verify-email [post]
func (h *AuthHandlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := commands.VerifyEmailCommand{
		Token: req.Token,
	}

	err := h.verifyEmailHandler.Handle(r.Context(), cmd)
	if err != nil {
		if err == appErrors.ErrInvalidInput {
			respondError(w, http.StatusBadRequest, "Invalid or expired verification token")
			return
		}
		if err == appErrors.ErrAlreadyExists {
			respondError(w, http.StatusConflict, "Email already verified")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify email")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Email verified successfully",
	})
}

// ResendVerification godoc
// @Summary Resend verification email
// @Description Resend the email verification link to the user's email address
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResendVerificationRequest true "User email"
// @Success 200 {object} map[string]string "Verification email sent"
// @Failure 400 {object} ErrorResponse "Invalid email"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 409 {object} ErrorResponse "Email already verified"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/resend-verification [post]
func (h *AuthHandlers) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := commands.ResendVerificationEmailCommand{
		Email: req.Email,
	}

	err := h.resendVerificationEmailHandler.Handle(r.Context(), cmd)
	if err != nil {
		if err == appErrors.ErrInvalidInput {
			respondError(w, http.StatusBadRequest, "Invalid email")
			return
		}
		if err == appErrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		if err == appErrors.ErrAlreadyExists {
			respondError(w, http.StatusConflict, "Email already verified")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to send verification email")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Verification email sent successfully",
	})
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Request a password reset email with a reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "User email"
// @Success 200 {object} map[string]string "Reset email sent successfully"
// @Failure 400 {object} ErrorResponse "Invalid email"
// @Failure 403 {object} ErrorResponse "Email not verified"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/forgot-password [post]
func (h *AuthHandlers) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := commands.RequestPasswordResetCommand{
		Email: req.Email,
	}

	err := h.requestPasswordResetHandler.Handle(r.Context(), cmd)
	if err != nil {
		if err == appErrors.ErrInvalidInput {
			respondError(w, http.StatusBadRequest, "Invalid email")
			return
		}
		if err == appErrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		if err == appErrors.ErrEmailNotVerified {
			respondError(w, http.StatusForbidden, "Please verify your email before resetting password")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to send reset email")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset email sent successfully",
	})
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset user password using the reset token from email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} map[string]string "Password reset successfully"
// @Failure 400 {object} ErrorResponse "Invalid token or password requirements not met"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/reset-password [post]
func (h *AuthHandlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := commands.ResetPasswordCommand{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}

	err := h.resetPasswordHandler.Handle(r.Context(), cmd)
	if err != nil {
		if err == appErrors.ErrInvalidInput {
			respondError(w, http.StatusBadRequest, "Invalid or expired token, or password requirements not met")
			return
		}
		if err == appErrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to reset password")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}
