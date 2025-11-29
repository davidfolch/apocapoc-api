package i18n

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/text/language"
)

//go:embed locales/en.json
var enTranslations []byte

//go:embed locales/es.json
var esTranslations []byte

type Translations struct {
	Errors     map[string]string `json:"errors"`
	Success    map[string]string `json:"success"`
	Validation map[string]string `json:"validation"`
	Emails     map[string]string `json:"emails"`
}

type Translator struct {
	translations map[language.Tag]Translations
	matcher      language.Matcher
}

func NewTranslator() (*Translator, error) {
	var enTrans, esTrans Translations

	if err := json.Unmarshal(enTranslations, &enTrans); err != nil {
		return nil, fmt.Errorf("failed to load English translations: %w", err)
	}

	if err := json.Unmarshal(esTranslations, &esTrans); err != nil {
		return nil, fmt.Errorf("failed to load Spanish translations: %w", err)
	}

	translations := map[language.Tag]Translations{
		language.English: enTrans,
		language.Spanish: esTrans,
	}

	matcher := language.NewMatcher([]language.Tag{
		language.English,
		language.Spanish,
	})

	return &Translator{
		translations: translations,
		matcher:      matcher,
	}, nil
}

func (t *Translator) GetLanguage(acceptLanguage string) language.Tag {
	if acceptLanguage == "" {
		return language.English
	}

	tags, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil || len(tags) == 0 {
		return language.English
	}

	_, index, _ := t.matcher.Match(tags...)
	supportedTags := []language.Tag{language.English, language.Spanish}
	if index < len(supportedTags) {
		return supportedTags[index]
	}

	return language.English
}

func (t *Translator) Translate(lang language.Tag, category, key string) string {
	trans, ok := t.translations[lang]
	if !ok {
		trans = t.translations[language.English]
	}

	var categoryMap map[string]string
	switch category {
	case "errors":
		categoryMap = trans.Errors
	case "success":
		categoryMap = trans.Success
	case "validation":
		categoryMap = trans.Validation
	case "emails":
		categoryMap = trans.Emails
	default:
		return key
	}

	if value, ok := categoryMap[key]; ok {
		return value
	}

	return key
}

func (t *Translator) Error(lang language.Tag, key string) string {
	return t.Translate(lang, "errors", key)
}

func (t *Translator) Success(lang language.Tag, key string) string {
	return t.Translate(lang, "success", key)
}

func (t *Translator) Validation(lang language.Tag, key string) string {
	return t.Translate(lang, "validation", key)
}

func (t *Translator) Email(lang language.Tag, key string) string {
	return t.Translate(lang, "emails", key)
}

func (t *Translator) TranslateValidationError(lang language.Tag, field, validationKey string) string {
	message := t.Validation(lang, validationKey)
	return strings.ReplaceAll(message, field, field)
}
