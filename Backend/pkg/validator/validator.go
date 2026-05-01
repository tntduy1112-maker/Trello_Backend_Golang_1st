package validator

import (
	"regexp"
	"strings"
	"unicode"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, err := range e {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Field)
		sb.WriteString(": ")
		sb.WriteString(err.Message)
	}
	return sb.String()
}

func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

type Validator struct {
	errors ValidationErrors
}

func New() *Validator {
	return &Validator{errors: make(ValidationErrors, 0)}
}

func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{Field: field, Message: message})
}

func (v *Validator) Required(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "is required")
	}
	return v
}

func (v *Validator) Email(field, value string) *Validator {
	if value == "" {
		return v
	}
	if !emailRegex.MatchString(value) {
		v.AddError(field, "must be a valid email address")
	}
	return v
}

func (v *Validator) MinLength(field, value string, min int) *Validator {
	if value == "" {
		return v
	}
	if len(value) < min {
		v.AddError(field, "must be at least "+itoa(min)+" characters")
	}
	return v
}

func (v *Validator) MaxLength(field, value string, max int) *Validator {
	if value == "" {
		return v
	}
	if len(value) > max {
		v.AddError(field, "must be at most "+itoa(max)+" characters")
	}
	return v
}

func (v *Validator) Password(field, value string) *Validator {
	if value == "" {
		return v
	}
	if len(value) < 8 {
		v.AddError(field, "must be at least 8 characters")
		return v
	}
	var hasUpper, hasLower, hasDigit bool
	for _, c := range value {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		v.AddError(field, "must contain uppercase, lowercase, and digit")
	}
	return v
}

func (v *Validator) OTP(field, value string) *Validator {
	if value == "" {
		v.AddError(field, "is required")
		return v
	}
	if len(value) != 6 {
		v.AddError(field, "must be 6 digits")
		return v
	}
	for _, c := range value {
		if !unicode.IsDigit(c) {
			v.AddError(field, "must contain only digits")
			return v
		}
	}
	return v
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
