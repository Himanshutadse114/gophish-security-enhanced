package middleware

import (
	"encoding/gob"
	"net/http"
	"sync"
	"time"

	"github.com/gophish/gophish/auth"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

// SessionTimeout defines the inactivity timeout duration (15 minutes)
const SessionTimeout = 15 * time.Minute

// SessionInvalidation tracks invalidated sessions server-side
// This prevents replay of session cookies even if they were captured before logout
type SessionInvalidation struct {
	mu              sync.RWMutex
	invalidatedSessions map[string]time.Time // Maps sessionID to invalidation time
}

var sessionInvalidation = &SessionInvalidation{
	invalidatedSessions: make(map[string]time.Time),
}

// InvalidateSession marks a session as invalidated (called on logout)
func InvalidateSession(sessionID string) {
	sessionInvalidation.mu.Lock()
	defer sessionInvalidation.mu.Unlock()
	sessionInvalidation.invalidatedSessions[sessionID] = time.Now()
	
	// Clean up old invalidations (older than 24 hours) to prevent memory leak
	cutoff := time.Now().Add(-24 * time.Hour)
	for sid, invTime := range sessionInvalidation.invalidatedSessions {
		if invTime.Before(cutoff) {
			delete(sessionInvalidation.invalidatedSessions, sid)
		}
	}
}

// IsSessionInvalidated checks if a session has been invalidated
func IsSessionInvalidated(sessionID string) bool {
	sessionInvalidation.mu.RLock()
	defer sessionInvalidation.mu.RUnlock()
	_, exists := sessionInvalidation.invalidatedSessions[sessionID]
	return exists
}

// GenerateSessionToken creates a unique session token
// This token is generated once on login and stored in the session cookie
func GenerateSessionToken() string {
	// Generate a secure random token (64 hex characters = 32 bytes)
	return auth.GenerateSecureKey(32)
}

// init registers the necessary models to be saved in the session later
func init() {
	gob.Register(&models.User{})
	gob.Register(&models.Flash{})
	gob.Register(time.Time{})
	Store.Options.HttpOnly = true
	// Set MaxAge to match SessionTimeout (15 minutes) to ensure cookies
	// expire properly and align with the inactivity timeout logic
	// This prevents stale sessions from persisting beyond the timeout period
	Store.Options.MaxAge = int(SessionTimeout.Seconds())
	// SameSite=Strict provides additional CSRF protection by preventing
	// cookies from being sent in cross-site requests
	Store.Options.SameSite = http.SameSiteStrictMode
}

// Store contains the session information for the request
// 
// SECURITY NOTE: The session keys are randomly generated on each application
// restart, which means all existing sessions become invalid after a restart.
// For production deployments, consider:
// 1. Loading keys from environment variables or configuration
// 2. Using persistent keys stored securely (e.g., in a secrets manager)
// 3. Implementing key rotation with backward compatibility
//
// Example for production:
//   signingKey := os.Getenv("SESSION_SIGNING_KEY")
//   encryptionKey := os.Getenv("SESSION_ENCRYPTION_KEY")
//   if len(signingKey) == 0 || len(encryptionKey) == 0 {
//       log.Fatal("Session keys must be set via environment variables")
//   }
var Store = sessions.NewCookieStore(
	[]byte(securecookie.GenerateRandomKey(64)), //Signing key
	[]byte(securecookie.GenerateRandomKey(32)))  //Encryption key
