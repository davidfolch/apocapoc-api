package i18n

import (
	"testing"

	"golang.org/x/text/language"
)

func TestNewTranslator(t *testing.T) {
	translator, err := NewTranslator()
	if err != nil {
		t.Fatalf("Failed to create translator: %v", err)
	}

	if translator == nil {
		t.Fatal("Expected translator to be non-nil")
	}

	if translator.translations == nil {
		t.Fatal("Expected translations map to be initialized")
	}

	if len(translator.translations) != 2 {
		t.Errorf("Expected 2 languages, got %d", len(translator.translations))
	}
}

func TestGetLanguage(t *testing.T) {
	translator, _ := NewTranslator()

	tests := []struct {
		name           string
		acceptLanguage string
		expected       language.Tag
	}{
		{
			name:           "English",
			acceptLanguage: "en-US",
			expected:       language.English,
		},
		{
			name:           "Spanish",
			acceptLanguage: "es-ES",
			expected:       language.Spanish,
		},
		{
			name:           "Empty defaults to English",
			acceptLanguage: "",
			expected:       language.English,
		},
		{
			name:           "Unknown language defaults to English",
			acceptLanguage: "fr-FR",
			expected:       language.English,
		},
		{
			name:           "Spanish with quality",
			acceptLanguage: "es-ES,es;q=0.9",
			expected:       language.Spanish,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := translator.GetLanguage(tt.acceptLanguage)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestError(t *testing.T) {
	translator, _ := NewTranslator()

	tests := []struct {
		name     string
		lang     language.Tag
		key      string
		expected string
	}{
		{
			name:     "English error message",
			lang:     language.English,
			key:      "invalid_request_body",
			expected: "Invalid request body",
		},
		{
			name:     "Spanish error message",
			lang:     language.Spanish,
			key:      "invalid_request_body",
			expected: "Cuerpo de solicitud inválido",
		},
		{
			name:     "Missing key returns key",
			lang:     language.English,
			key:      "non_existent_key",
			expected: "non_existent_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := translator.Error(tt.lang, tt.key)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSuccess(t *testing.T) {
	translator, _ := NewTranslator()

	tests := []struct {
		name     string
		lang     language.Tag
		key      string
		expected string
	}{
		{
			name:     "English success message",
			lang:     language.English,
			key:      "logged_out",
			expected: "Successfully logged out",
		},
		{
			name:     "Spanish success message",
			lang:     language.Spanish,
			key:      "logged_out",
			expected: "Sesión cerrada exitosamente",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := translator.Success(tt.lang, tt.key)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestValidation(t *testing.T) {
	translator, _ := NewTranslator()

	tests := []struct {
		name     string
		lang     language.Tag
		key      string
		expected string
	}{
		{
			name:     "English validation message",
			lang:     language.English,
			key:      "email_required",
			expected: "email is required",
		},
		{
			name:     "Spanish validation message",
			lang:     language.Spanish,
			key:      "email_required",
			expected: "el email es requerido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := translator.Validation(tt.lang, tt.key)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestEmail(t *testing.T) {
	translator, _ := NewTranslator()

	tests := []struct {
		name     string
		lang     language.Tag
		key      string
		expected string
	}{
		{
			name:     "English email message",
			lang:     language.English,
			key:      "welcome_subject",
			expected: "Welcome to Apocapoc!",
		},
		{
			name:     "Spanish email message",
			lang:     language.Spanish,
			key:      "welcome_subject",
			expected: "¡Bienvenido a Apocapoc!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := translator.Email(tt.lang, tt.key)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestTranslate(t *testing.T) {
	translator, _ := NewTranslator()

	tests := []struct {
		name     string
		lang     language.Tag
		category string
		key      string
		expected string
	}{
		{
			name:     "Valid category and key",
			lang:     language.English,
			category: "errors",
			key:      "user_not_found",
			expected: "User not found",
		},
		{
			name:     "Invalid category returns key",
			lang:     language.English,
			category: "invalid_category",
			key:      "some_key",
			expected: "some_key",
		},
		{
			name:     "Unsupported language fallback to English",
			lang:     language.French,
			category: "errors",
			key:      "user_not_found",
			expected: "User not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := translator.Translate(tt.lang, tt.category, tt.key)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
