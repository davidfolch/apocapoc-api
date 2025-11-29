package i18n

import (
	"context"
	"net/http"

	"golang.org/x/text/language"
)

type contextKey string

const languageKey contextKey = "language"

func LanguageMiddleware(translator *Translator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acceptLanguage := r.Header.Get("Accept-Language")
			lang := translator.GetLanguage(acceptLanguage)
			ctx := context.WithValue(r.Context(), languageKey, lang)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetLanguageFromContext(ctx context.Context) language.Tag {
	if lang, ok := ctx.Value(languageKey).(language.Tag); ok {
		return lang
	}
	return language.English
}
