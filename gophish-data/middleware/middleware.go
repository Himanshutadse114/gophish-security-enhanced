package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
)

// CSRFExemptPrefixes are a list of routes that are exempt from CSRF protection
var CSRFExemptPrefixes = []string{
	"/api",
}

// CSRFExceptions is a middleware that prevents CSRF checks on routes listed in
// CSRFExemptPrefixes.
func CSRFExceptions(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, prefix := range CSRFExemptPrefixes {
			if strings.HasPrefix(r.URL.Path, prefix) {
				r = csrf.UnsafeSkipCheck(r)
				break
			}
		}
		handler.ServeHTTP(w, r)
	}
}

// Use allows us to stack middleware to process the request
// Example taken from https://github.com/gorilla/mux/pull/36#issuecomment-25849172
func Use(handler http.HandlerFunc, mid ...func(http.Handler) http.HandlerFunc) http.HandlerFunc {
	for _, m := range mid {
		handler = m(handler)
	}
	return handler
}

// GetContext wraps each request in a function which fills in the context for a given request.
// This includes setting the User and Session keys and values as necessary for use in later functions.
func GetContext(handler http.Handler) http.HandlerFunc {
	// Set the context here
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request form
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing request", http.StatusInternalServerError)
		}
		// Set the context appropriately here.
		// Set the session
		session, err := Store.Get(r, "gophish")
		if err != nil {
			// Log session retrieval errors but continue with a new session
			// This prevents session key rotation issues from breaking the app
			log.Warnf("Error retrieving session: %v. Creating new session.", err)
			session, _ = Store.New(r, "gophish")
		}
		
		// Security: If session exists but has MaxAge=-1 (from logout), treat as invalid
		// This prevents replay of old session cookies after logout
		if session.Options.MaxAge == -1 {
			// Session was invalidated (e.g., via logout), create a fresh session
			session, _ = Store.New(r, "gophish")
		}
		
		// Put the session in the context so that we can
		// reuse the values in different handlers
		r = ctx.Set(r, "session", session)
		
		// Security: Check if session has been invalidated server-side (prevents replay attacks)
		if sessionToken, ok := session.Values["sessionToken"]; ok {
			if token, ok := sessionToken.(string); ok {
				if IsSessionInvalidated(token) {
					// Session was invalidated (e.g., via logout), clear it
					delete(session.Values, "id")
					delete(session.Values, "lastActivity")
					delete(session.Values, "passwordVersion")
					delete(session.Values, "sessionToken")
					session.Save(r, w)
					r = ctx.Set(r, "user", nil)
					handler.ServeHTTP(w, r)
					ctx.Clear(r)
					return
				}
			}
		}
		
		// Check for session timeout (15 minutes of inactivity)
		if id, ok := session.Values["id"]; ok {
			// Check if lastActivity exists in session
			if lastActivity, ok := session.Values["lastActivity"]; ok {
				lastActivityTime, ok := lastActivity.(time.Time)
				if !ok {
					// Invalid time format, clear session
					delete(session.Values, "id")
					delete(session.Values, "lastActivity")
					delete(session.Values, "passwordVersion")
					session.Save(r, w)
					r = ctx.Set(r, "user", nil)
				} else {
					// Check if session has expired (15 minutes of inactivity)
					if time.Since(lastActivityTime) > SessionTimeout {
						// Session expired, clear it
						delete(session.Values, "id")
						delete(session.Values, "lastActivity")
						delete(session.Values, "passwordVersion")
						session.Save(r, w)
						r = ctx.Set(r, "user", nil)
					} else {
						// Session is still valid, get user and update activity time
						u, err := models.GetUser(id.(int64))
						if err != nil {
							r = ctx.Set(r, "user", nil)
						} else {
							// Security: Check if password version matches
							// This invalidates all sessions when password is changed
							sessionPasswordVersion, sessionHasVersion := session.Values["passwordVersion"]
							if !sessionHasVersion {
								// Old session without version, invalidate it
								delete(session.Values, "id")
								delete(session.Values, "lastActivity")
								delete(session.Values, "passwordVersion")
								session.Save(r, w)
								r = ctx.Set(r, "user", nil)
							} else {
								// Compare session version with user's current version
								sessionVersion, ok := sessionPasswordVersion.(int64)
								if !ok || sessionVersion != u.PasswordVersion {
									// Password was changed, invalidate this session
									delete(session.Values, "id")
									delete(session.Values, "lastActivity")
									delete(session.Values, "passwordVersion")
									session.Save(r, w)
									r = ctx.Set(r, "user", nil)
								} else {
									// Session is valid, set user and update activity time
									r = ctx.Set(r, "user", u)
									// Update last activity time
									session.Values["lastActivity"] = time.Now()
									session.Save(r, w)
								}
							}
						}
					}
				}
			} else {
				// First time or no lastActivity set, initialize it
				u, err := models.GetUser(id.(int64))
				if err != nil {
					r = ctx.Set(r, "user", nil)
				} else {
					// Check password version for old sessions
					sessionPasswordVersion, sessionHasVersion := session.Values["passwordVersion"]
					if !sessionHasVersion {
						// Old session without version, set it
						session.Values["passwordVersion"] = u.PasswordVersion
					} else {
						// Verify version matches
						sessionVersion, ok := sessionPasswordVersion.(int64)
						if !ok || sessionVersion != u.PasswordVersion {
							// Password was changed, invalidate this session
							delete(session.Values, "id")
							delete(session.Values, "lastActivity")
							delete(session.Values, "passwordVersion")
							session.Save(r, w)
							r = ctx.Set(r, "user", nil)
							handler.ServeHTTP(w, r)
							ctx.Clear(r)
							return
						}
					}
					r = ctx.Set(r, "user", u)
					session.Values["lastActivity"] = time.Now()
					session.Save(r, w)
				}
			}
		} else {
			r = ctx.Set(r, "user", nil)
		}
		handler.ServeHTTP(w, r)
		// Remove context contents
		ctx.Clear(r)
	}
}

// RequireAPIKey ensures that a valid API key is set as either the api_key GET
// parameter, or a Bearer token. Additionally, it requires a valid session and
// validates that the token belongs to the current session user to prevent IDOR attacks.
// This ensures API keys cannot be used after logout or session expiration.
func RequireAPIKey(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Max-Age", "1000")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			return
		}
		// Security: Require a valid session for API access
		// This ensures API keys cannot be used after logout or session expiration
		sessionVal := ctx.Get(r, "session")
		if sessionVal == nil {
			JSONError(w, http.StatusUnauthorized, "Session required for API access")
			return
		}
		
		session := sessionVal.(*sessions.Session)
		
		// Strict validation: Check that session has all required values
		sessionUserID, hasSessionID := session.Values["id"]
		if !hasSessionID {
			JSONError(w, http.StatusUnauthorized, "Valid session required for API access")
			return
		}
		
		// Security: Check if session has been invalidated server-side (prevents replay attacks)
		if sessionToken, ok := session.Values["sessionToken"]; ok {
			if token, ok := sessionToken.(string); ok {
				if IsSessionInvalidated(token) {
					JSONError(w, http.StatusUnauthorized, "Session has been invalidated")
					return
				}
			}
		}
		
		// Additional security: Verify session has lastActivity (prevents use of stale sessions)
		lastActivity, hasLastActivity := session.Values["lastActivity"]
		if !hasLastActivity {
			JSONError(w, http.StatusUnauthorized, "Invalid session state")
			return
		}
		
		// Verify lastActivity is a valid time and session hasn't expired
		lastActivityTime, ok := lastActivity.(time.Time)
		if !ok {
			JSONError(w, http.StatusUnauthorized, "Invalid session state")
			return
		}
		
		// Check if session has expired (15 minutes of inactivity)
		if time.Since(lastActivityTime) > SessionTimeout {
			JSONError(w, http.StatusUnauthorized, "Session expired")
			return
		}
		
		// Verify session user exists and is valid (GetContext middleware should have set this)
		sessionUser := ctx.Get(r, "user")
		if sessionUser == nil {
			JSONError(w, http.StatusUnauthorized, "Valid session user required for API access")
			return
		}
		
		// Parse API key from request
		r.ParseForm()
		ak := r.Form.Get("api_key")
		// If we can't get the API key, we'll also check for the
		// Authorization Bearer token
		if ak == "" {
			tokens, ok := r.Header["Authorization"]
			if ok && len(tokens) >= 1 {
				ak = tokens[0]
				ak = strings.TrimPrefix(ak, "Bearer ")
			}
		}
		if ak == "" {
			JSONError(w, http.StatusUnauthorized, "API Key not set")
			return
		}
		
		// Validate API key
		u, err := models.GetUserByAPIKey(ak)
		if err != nil {
			JSONError(w, http.StatusUnauthorized, "Invalid API Key")
			return
		}
		
		// Security: API key user must match session user
		// This prevents users from using another user's token and ensures
		// API keys are tied to active sessions
		sessionUID := sessionUserID.(int64)
		if u.Id != sessionUID {
			JSONError(w, http.StatusForbidden, "API key does not belong to current session user")
			return
		}
		
		// Additional validation: Ensure session user matches API key user
		sessionUserObj := sessionUser.(models.User)
		if u.Id != sessionUserObj.Id {
			JSONError(w, http.StatusForbidden, "API key user mismatch with session user")
			return
		}
		
		r = ctx.Set(r, "user", u)
		r = ctx.Set(r, "user_id", u.Id)
		r = ctx.Set(r, "api_key", ak)
		handler.ServeHTTP(w, r)
	})
}

// RequireLogin checks to see if the user is currently logged in.
// If not, the function returns a 302 redirect to the login page.
func RequireLogin(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if u := ctx.Get(r, "user"); u != nil {
			// If a password change is required for the user, then redirect them
			// to the login page
			currentUser := u.(models.User)
			if currentUser.PasswordChangeRequired && r.URL.Path != "/reset_password" {
				q := r.URL.Query()
				q.Set("next", r.URL.Path)
				http.Redirect(w, r, fmt.Sprintf("/reset_password?%s", q.Encode()), http.StatusTemporaryRedirect)
				return
			}
			handler.ServeHTTP(w, r)
			return
		}
		q := r.URL.Query()
		q.Set("next", r.URL.Path)
		http.Redirect(w, r, fmt.Sprintf("/login?%s", q.Encode()), http.StatusTemporaryRedirect)
	}
}

// EnforceViewOnly is a global middleware that limits the ability to edit
// objects to accounts with the PermissionModifyObjects permission.
func EnforceViewOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the request is for any non-GET HTTP method, e.g. POST, PUT,
		// or DELETE, we need to ensure the user has the appropriate
		// permission.
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			user := ctx.Get(r, "user").(models.User)
			access, err := user.HasPermission(models.PermissionModifyObjects)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			if !access {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePermission checks to see if the user has the requested permission
// before executing the handler. If the request is unauthorized, a JSONError
// is returned.
func RequirePermission(perm string) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user := ctx.Get(r, "user").(models.User)
			access, err := user.HasPermission(perm)
			if err != nil {
				JSONError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if !access {
				JSONError(w, http.StatusForbidden, http.StatusText(http.StatusForbidden))
				return
			}
			next.ServeHTTP(w, r)
		}
	}
}

// ApplySecurityHeaders applies various security headers according to best-
// practices. This includes protection against XSS, clickjacking, MIME sniffing,
// and enforces secure transport when using HTTPS.
func ApplySecurityHeaders(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// X-Content-Type-Options: Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		
		// X-XSS-Protection: Enable XSS filtering (legacy browsers)
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// X-Frame-Options: Prevent clickjacking attacks
		w.Header().Set("X-Frame-Options", "DENY")
		
		// Strict-Transport-Security: Enforce HTTPS connections
		// Only set on HTTPS connections (check scheme or X-Forwarded-Proto header)
		isHTTPS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" || r.URL.Scheme == "https"
		if isHTTPS {
			// max-age=63072000 = 2 years, includeSubDomains, preload
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}
		
		// Content-Security-Policy: Restrict resource loading
		// Using a balanced policy that maintains functionality while providing security
		// Note: For initial deployment, consider using Content-Security-Policy-Report-Only
		// to test without breaking functionality. Once validated, switch to enforcement mode.
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " + // unsafe-inline/unsafe-eval needed for some frameworks
			"style-src 'self' 'unsafe-inline'; " + // unsafe-inline needed for inline styles
			"img-src 'self' data: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self'; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'; " +
			"frame-ancestors 'none';"
		w.Header().Set("Content-Security-Policy", csp)
		
		// Referrer-Policy: Control referrer information
		w.Header().Set("Referrer-Policy", "no-referrer")
		
		// Permissions-Policy: Restrict browser features
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		
		next.ServeHTTP(w, r)
	}
}

// JSONError returns an error in JSON format with the given
// status code and message
func JSONError(w http.ResponseWriter, c int, m string) {
	cj, _ := json.MarshalIndent(models.Response{Success: false, Message: m}, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	fmt.Fprintf(w, "%s", cj)
}
