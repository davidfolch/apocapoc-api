package http

import (
	"net/http"

	"apocapoc-api/internal/application/commands"
	"apocapoc-api/internal/i18n"
	"apocapoc-api/internal/shared/errors"
)

type UserHandlers struct {
	deleteUserHandler *commands.DeleteUserHandler
	translator        *i18n.Translator
}

func NewUserHandlers(deleteUserHandler *commands.DeleteUserHandler, translator *i18n.Translator) *UserHandlers {
	return &UserHandlers{
		deleteUserHandler: deleteUserHandler,
		translator:        translator,
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
			respondErrorI18n(w, r, h.translator, http.StatusNotFound, "user_not_found")
			return
		}
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "failed_delete_user")
		return
	}

	lang := i18n.GetLanguageFromContext(r.Context())
	respondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.Success(lang, "user_deleted"),
	})
}
