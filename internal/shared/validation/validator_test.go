package validation

import (
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "user@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"valid email with plus", "user+tag@example.com", false},
		{"valid email with dots", "user.name@example.com", false},
		{"valid email with numbers", "user123@example.com", false},
		{"valid email with dash", "user-name@example.com", false},
		{"empty email", "", true},
		{"missing @", "userexample.com", true},
		{"missing domain", "user@", true},
		{"missing local part", "@example.com", true},
		{"invalid format", "string", true},
		{"double @", "user@@example.com", true},
		{"spaces in email", "user name@example.com", true},
		{"too long email", strings.Repeat("a", 250) + "@example.com", true},
		{"too long local part", strings.Repeat("a", 65) + "@example.com", true},
		{"no TLD", "user@example", false},
		{"with whitespace", "  user@example.com  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		pwd     string
		wantErr bool
		errMsg  string
	}{
		{"valid strong password", "Passw0rd!", false, ""},
		{"valid with symbols", "MyP@ssw0rd#2024", false, ""},
		{"valid with mixed case", "Str0ng!Pass", false, ""},
		{"empty password", "", true, "password is required"},
		{"too short", "Pass1!", true, "at least 8 characters"},
		{"no uppercase", "password1!", true, "uppercase letter"},
		{"no lowercase", "PASSWORD1!", true, "lowercase letter"},
		{"no digit", "Password!", true, "digit"},
		{"no special char", "Password1", true, "special character"},
		{"only letters", "PasswordPassword", true, "digit"},
		{"only numbers", "12345678", true, "uppercase letter"},
		{"7 chars valid format", "Passw0!", true, "at least 8 characters"},
		{"exactly 8 chars", "Passw0rd!", false, ""},
		{"very long password", strings.Repeat("Aa1!", 32), false, ""},
		{"too long password", strings.Repeat("a", 129), true, "must not exceed 128 characters"},
		{"unicode special chars", "Pässw0rd!", false, ""},
		{"spaces do not count as special", "Pass word1", true, "special character"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.pwd)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tt.pwd, err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePassword(%q) error = %v, want error containing %q", tt.pwd, err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		wantErr  bool
	}{
		{"valid UTC", "UTC", false},
		{"valid America/New_York", "America/New_York", false},
		{"valid Europe/Madrid", "Europe/Madrid", false},
		{"valid Asia/Tokyo", "Asia/Tokyo", false},
		{"valid Europe/London", "Europe/London", false},
		{"valid Australia/Sydney", "Australia/Sydney", false},
		{"valid with spaces trimmed", "  UTC  ", false},
		{"empty timezone", "", true},
		{"invalid timezone", "string", true},
		{"invalid format", "Invalid/Timezone", true},
		{"numeric timezone", "GMT+1", true},
		{"partial timezone", "America", true},
		{"lowercase valid", "utc", true},
		{"typo in timezone", "America/New_Yorkkk", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimezone(tt.timezone)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimezone(%q) error = %v, wantErr %v", tt.timezone, err, tt.wantErr)
			}
		})
	}
}

func TestValidateRegistration(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{
			"valid registration",
			"user@example.com",
			"Passw0rd!",
			false,
		},
		{
			"valid with complex email",
			"user.name+tag@example.co.uk",
			"MyS3cur3P@ss",
			false,
		},
		{
			"invalid email",
			"invalid-email",
			"Passw0rd!",
			true,
		},
		{
			"invalid password",
			"user@example.com",
			"weak",
			true,
		},
		{
			"all invalid",
			"not-an-email",
			"weak",
			true,
		},
		{
			"empty email",
			"",
			"Passw0rd!",
			true,
		},
		{
			"empty password",
			"user@example.com",
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistration(tt.email, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRegistration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "email",
		Message: "invalid format",
	}

	expected := "email: invalid format"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), expected)
	}
}
