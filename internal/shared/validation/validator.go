package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// RFC 5322 compliant email regex (simplified but robust)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return ValidationError{Field: "email", Message: "email is required"}
	}

	if len(email) > 254 {
		return ValidationError{Field: "email", Message: "email must not exceed 254 characters"}
	}

	if !emailRegex.MatchString(email) {
		return ValidationError{Field: "email", Message: "invalid email format"}
	}

	parts := strings.Split(email, "@")
	if len(parts[0]) > 64 {
		return ValidationError{Field: "email", Message: "email local part must not exceed 64 characters"}
	}

	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return ValidationError{Field: "password", Message: "password is required"}
	}

	if len(password) < 8 {
		return ValidationError{Field: "password", Message: "password must be at least 8 characters long"}
	}

	if len(password) > 128 {
		return ValidationError{Field: "password", Message: "password must not exceed 128 characters"}
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return ValidationError{Field: "password", Message: "password must contain at least one uppercase letter"}
	}

	if !hasLower {
		return ValidationError{Field: "password", Message: "password must contain at least one lowercase letter"}
	}

	if !hasDigit {
		return ValidationError{Field: "password", Message: "password must contain at least one digit"}
	}

	if !hasSpecial {
		return ValidationError{Field: "password", Message: "password must contain at least one special character"}
	}

	if IsCommonPassword(password) {
		return ValidationError{Field: "password", Message: "password is too common, please choose a more secure password"}
	}

	return nil
}

func ValidateTimezone(timezone string) error {
	timezone = strings.TrimSpace(timezone)

	if timezone == "" {
		return ValidationError{Field: "timezone", Message: "timezone is required"}
	}

	_, err := time.LoadLocation(timezone)
	if err != nil {
		return ValidationError{Field: "timezone", Message: "invalid timezone, must be a valid IANA timezone (e.g., 'America/New_York', 'Europe/Madrid', 'UTC')"}
	}

	return nil
}

func ValidateRegistration(email, password string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	if err := ValidatePassword(password); err != nil {
		return err
	}

	return nil
}
