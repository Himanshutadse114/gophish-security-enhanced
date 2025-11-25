package middleware

import (
	"os"
	"strconv"
	"time"
)

// GetSessionTimeout returns the session timeout duration.
// For testing purposes, this can be overridden via SESSION_TIMEOUT_MINUTES environment variable.
// Default is 15 minutes.
func GetSessionTimeout() time.Duration {
	timeoutStr := os.Getenv("SESSION_TIMEOUT_MINUTES")
	if timeoutStr != "" {
		if minutes, err := strconv.Atoi(timeoutStr); err == nil && minutes > 0 {
			return time.Duration(minutes) * time.Minute
		}
	}
	return SessionTimeout
}

// IsTestMode returns true if running in test mode (timeout < 5 minutes)
// This can be used to add additional logging or behavior during testing
func IsTestMode() bool {
	return GetSessionTimeout() < 5*time.Minute
}

