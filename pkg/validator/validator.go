package validator

import (
	"fmt"
	"net/mail"
	"regexp"

	"github.com/RobinHood3082/simplebank/util"
)

var (
	isValidUsername = regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString
	isValidFullName = regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString
)

func ValidateString(value string, minLength int, maxLength int) error {
	n := len(value)
	if n < minLength || n > maxLength {
		return fmt.Errorf("must contain from %d to %d characters", minLength, maxLength)
	}
	return nil
}

func ValidateUsername(value string) error {
	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}

	if !isValidUsername(value) {
		return fmt.Errorf("must contain only letters, numbers, and underscores")
	}
	return nil
}

func ValidatePassword(value string) error {
	return ValidateString(value, 6, 100)
}

func ValidateEmail(value string) error {
	if err := ValidateString(value, 3, 200); err != nil {
		return err
	}

	_, err := mail.ParseAddress(value)
	if err != nil {
		return fmt.Errorf("must be a valid email address")
	}
	return nil
}

func ValidateFullName(value string) error {
	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}

	if !isValidFullName(value) {
		return fmt.Errorf("must contain only letters and spaces")
	}
	return nil
}

func ValidateEmailId(value int64) error {
	if value <= 0 {
		return fmt.Errorf("must be a positive integer")
	}
	return nil
}

func ValidateVerifyEmailCode(value string) error {
	if err := ValidateString(value, 32, 128); err != nil {
		return err
	}
	return nil
}

func ValidateCurrency(currency string) error {
	if !util.IsCurrencySupported(currency) {
		return fmt.Errorf("currency is not supported")
	}
	return nil
}
