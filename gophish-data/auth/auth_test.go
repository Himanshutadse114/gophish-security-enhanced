package auth

import (
	"testing"
)

func TestPasswordPolicy(t *testing.T) {
	// Test empty password
	candidate := ""
	got := CheckPasswordPolicy(candidate)
	if got != ErrEmptyPassword {
		t.Fatalf("unexpected error received. expected %v got %v", ErrEmptyPassword, got)
	}

	// Test too short password
	candidate = "short"
	got = CheckPasswordPolicy(candidate)
	if got != ErrPasswordTooShort {
		t.Fatalf("unexpected error received. expected %v got %v", ErrPasswordTooShort, got)
	}

	// Test password without uppercase
	candidate = "validpassword123!"
	got = CheckPasswordPolicy(candidate)
	if got != ErrPasswordNoUppercase {
		t.Fatalf("unexpected error received. expected %v got %v", ErrPasswordNoUppercase, got)
	}

	// Test password without lowercase
	candidate = "VALIDPASSWORD123!"
	got = CheckPasswordPolicy(candidate)
	if got != ErrPasswordNoLowercase {
		t.Fatalf("unexpected error received. expected %v got %v", ErrPasswordNoLowercase, got)
	}

	// Test password without number
	candidate = "ValidPassword!"
	got = CheckPasswordPolicy(candidate)
	if got != ErrPasswordNoNumber {
		t.Fatalf("unexpected error received. expected %v got %v", ErrPasswordNoNumber, got)
	}

	// Test password without special character
	candidate = "ValidPassword123"
	got = CheckPasswordPolicy(candidate)
	if got != ErrPasswordNoSpecialChar {
		t.Fatalf("unexpected error received. expected %v got %v", ErrPasswordNoSpecialChar, got)
	}

	// Test valid password
	candidate = "ValidPassword123!"
	got = CheckPasswordPolicy(candidate)
	if got != nil {
		t.Fatalf("unexpected error received. expected %v got %v", nil, got)
	}

	// Test valid password with different special characters
	candidate = "MyStr0ng@Pass#word"
	got = CheckPasswordPolicy(candidate)
	if got != nil {
		t.Fatalf("unexpected error received. expected %v got %v", nil, got)
	}
}

func TestValidatePasswordChange(t *testing.T) {
	currentPassword := "CurrentPassword123!"
	currentHash, err := GeneratePasswordHash(currentPassword)
	if err != nil {
		t.Fatalf("unexpected error generating password hash: %v", err)
	}

	// Test password mismatch
	newPassword := "NewValidPassword123!"
	confirmPassword := "invalid"
	_, got := ValidatePasswordChange(currentHash, newPassword, confirmPassword)
	if got != ErrPasswordMismatch {
		t.Fatalf("unexpected error received. expected %v got %v", ErrPasswordMismatch, got)
	}

	// Test password reuse
	newPassword = currentPassword
	confirmPassword = newPassword
	_, got = ValidatePasswordChange(currentHash, newPassword, confirmPassword)
	if got != ErrReusedPassword {
		t.Fatalf("unexpected error received. expected %v got %v", ErrReusedPassword, got)
	}

	// Test valid password change
	newPassword = "NewValidPassword123!"
	confirmPassword = newPassword
	_, got = ValidatePasswordChange(currentHash, newPassword, confirmPassword)
	if got != nil {
		t.Fatalf("unexpected error received. expected %v got %v", nil, got)
	}
}
