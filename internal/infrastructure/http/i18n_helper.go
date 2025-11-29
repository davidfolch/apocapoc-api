package http

import (
	"net/http"
	"strings"

	"apocapoc-api/internal/i18n"

	"golang.org/x/text/language"
)

func respondErrorI18n(w http.ResponseWriter, r *http.Request, translator *i18n.Translator, status int, key string) {
	lang := i18n.GetLanguageFromContext(r.Context())
	message := translator.Error(lang, key)
	respondJSON(w, status, ErrorResponse{Error: message})
}

func respondSuccessI18n(w http.ResponseWriter, r *http.Request, translator *i18n.Translator, key string) {
	lang := i18n.GetLanguageFromContext(r.Context())
	message := translator.Success(lang, key)
	respondJSON(w, http.StatusOK, map[string]string{"message": message})
}

func respondValidationErrorI18n(w http.ResponseWriter, r *http.Request, translator *i18n.Translator, err error) {
	lang := i18n.GetLanguageFromContext(r.Context())
	errMsg := err.Error()
	var field string
	var translatedMsg string

	if strings.Contains(errMsg, ": ") {
		parts := strings.SplitN(errMsg, ": ", 3)
		if len(parts) >= 3 {
			field = parts[1]
			validationKey := parts[2]
			translatedMsg = translator.Validation(lang, validationKey)
			respondJSON(w, http.StatusBadRequest, ValidationErrorResponse{
				Error: translatedMsg,
				Field: field,
			})
			return
		}
	}

	respondJSON(w, http.StatusBadRequest, ErrorResponse{
		Error: errMsg,
	})
}

func getLanguageFromRequest(r *http.Request, translator *i18n.Translator) language.Tag {
	lang := i18n.GetLanguageFromContext(r.Context())
	return lang
}
