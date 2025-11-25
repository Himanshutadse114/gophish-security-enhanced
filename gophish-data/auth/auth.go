package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// MinPasswordLength is the minimum number of characters required in a password
const MinPasswordLength = 12

// APIKeyLength is the length of Gophish API keys
const APIKeyLength = 32

// ErrInvalidPassword is thrown when a user provides an incorrect password.
var ErrInvalidPassword = errors.New("Invalid Password")

// ErrPasswordMismatch is thrown when a user provides a mismatching password
// and confirmation password.
var ErrPasswordMismatch = errors.New("Passwords do not match")

// ErrReusedPassword is thrown when a user attempts to change their password to
// the existing password
var ErrReusedPassword = errors.New("Cannot reuse existing password")

// ErrEmptyPassword is thrown when a user provides a blank password to the register
// or change password functions
var ErrEmptyPassword = errors.New("No password provided")

// ErrPasswordTooShort is thrown when a user provides a password that is less
// than MinPasswordLength
var ErrPasswordTooShort = fmt.Errorf("Password must be at least %d characters", MinPasswordLength)

// ErrPasswordNoUppercase is thrown when a password doesn't contain an uppercase letter
var ErrPasswordNoUppercase = errors.New("Password must contain at least one uppercase letter")

// ErrPasswordNoLowercase is thrown when a password doesn't contain a lowercase letter
var ErrPasswordNoLowercase = errors.New("Password must contain at least one lowercase letter")

// ErrPasswordNoNumber is thrown when a password doesn't contain a number
var ErrPasswordNoNumber = errors.New("Password must contain at least one number")

// ErrPasswordNoSpecialChar is thrown when a password doesn't contain a special character
var ErrPasswordNoSpecialChar = errors.New("Password must contain at least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)")

// GenerateSecureKey returns the hex representation of key generated from n
// random bytes
func GenerateSecureKey(n int) string {
	k := make([]byte, n)
	io.ReadFull(rand.Reader, k)
	return fmt.Sprintf("%x", k)
}

// GeneratePasswordHash returns the bcrypt hash for the provided password using
// the default bcrypt cost.
func GeneratePasswordHash(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// CheckPasswordPolicy ensures the provided password is valid according to our
// strong password policy.
//
// The password policy requires:
// - Minimum 12 characters
// - At least one uppercase letter (A-Z)
// - At least one lowercase letter (a-z)
// - At least one number (0-9)
// - At least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)
func CheckPasswordPolicy(password string) error {
	// Check for empty password
	if password == "" {
		return ErrEmptyPassword
	}
	
	// Check minimum length
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	
	// Check for required character types
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false
	
	// Define special characters
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case strings.ContainsRune(specialChars, char):
			hasSpecial = true
		}
	}
	
	// Return specific error for missing requirements
	if !hasUpper {
		return ErrPasswordNoUppercase
	}
	if !hasLower {
		return ErrPasswordNoLowercase
	}
	if !hasNumber {
		return ErrPasswordNoNumber
	}
	if !hasSpecial {
		return ErrPasswordNoSpecialChar
	}
	
	return nil
}

// ValidatePassword validates that the provided password matches the provided
// bcrypt hash.
func ValidatePassword(password string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ValidatePasswordChange validates that the new password matches the
// configured password policy, that the new password and confirmation
// password match.
//
// Note that this assumes the current password has been confirmed by the
// caller.
//
// If all of the provided data is valid, then the hash of the new password is
// returned.
func ValidatePasswordChange(currentHash, newPassword, confirmPassword string) (string, error) {
	// Ensure the new password passes our password policy
	if err := CheckPasswordPolicy(newPassword); err != nil {
		return "", err
	}
	// Check that new passwords match
	if newPassword != confirmPassword {
		return "", ErrPasswordMismatch
	}
	// Make sure that the new password isn't the same as the old one
	err := ValidatePassword(newPassword, currentHash)
	if err == nil {
		return "", ErrReusedPassword
	}
	// Generate the new hash
	return GeneratePasswordHash(newPassword)
}
