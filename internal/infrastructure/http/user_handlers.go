package http

import (
	"net/http"

	"apocapoc-api/internal/application/commands"
	"apocapoc-api/internal/shared/errors"
)

type UserHandlers struct {
	deleteUserHandler *commands.DeleteUserHandler
}

func NewUserHandlers(deleteUserHandler *commands.DeleteUserHandler) *UserHandlers {
	return &UserHandlers{
		deleteUserHandler: deleteUserHandler,
	}
}

// DeleteAccount godoc
// @Summary Delete user account
// @Description Permanently delete the authenticated user's account and all associated data (habits, entries, tokens). This action cannot be undone.
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string "Account deleted successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing token"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/me [delete]
func (h *UserHandlers) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	cmd := commands.DeleteUserCommand{
		UserID: userID,
	}

	err := h.deleteUserHandler.Handle(r.Context(), cmd)
	if err != nil {
		if err == errors.ErrNotFound {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Account deleted successfully",
	})
}
