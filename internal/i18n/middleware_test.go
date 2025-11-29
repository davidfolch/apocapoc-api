package i18n

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/text/language"
)

func TestLanguageMiddleware(t *testing.T) {
	translator, _ := NewTranslator()
	middleware := LanguageMiddleware(translator)

	tests := []struct {
		name           string
		acceptLanguage string
		expectedLang   language.Tag
	}{
		{
			name:           "English header",
			acceptLanguage: "en-US",
			expectedLang:   language.English,
		},
		{
			name:           "Spanish header",
			acceptLanguage: "es-ES",
			expectedLang:   language.Spanish,
		},
		{
			name:           "No header defaults to English",
			acceptLanguage: "",
			expectedLang:   language.English,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedLang language.Tag
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedLang = GetLanguageFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			wrappedHandler := middleware(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tt.acceptLanguage)
			}

			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)

			if capturedLang != tt.expectedLang {
				t.Errorf("Expected language %v, got %v", tt.expectedLang, capturedLang)
			}
		})
	}
}

func TestGetLanguageFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected language.Tag
	}{
		{
			name:     "Context with English",
			ctx:      context.WithValue(context.Background(), languageKey, language.English),
			expected: language.English,
		},
		{
			name:     "Context with Spanish",
			ctx:      context.WithValue(context.Background(), languageKey, language.Spanish),
			expected: language.Spanish,
		},
		{
			name:     "Context without language defaults to English",
			ctx:      context.Background(),
			expected: language.English,
		},
		{
			name:     "Context with wrong value type defaults to English",
			ctx:      context.WithValue(context.Background(), languageKey, "invalid"),
			expected: language.English,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLanguageFromContext(tt.ctx)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
