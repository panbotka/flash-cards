package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// SessionCookieName is the name of the session cookie.
	SessionCookieName = "flash_session"

	// SessionTTL is the lifetime of a session cookie (30 days).
	SessionTTL = 30 * 24 * time.Hour

	// sessionPayload is the fixed message signed by the HMAC.
	sessionPayload = "flash-cards-session"
)

// NewMiddleware returns a Gin middleware that enforces password-based session
// authentication. If password is empty (dev mode), the middleware is a no-op
// and all requests are allowed through.
func NewMiddleware(password string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Dev mode: no password configured, skip all auth checks.
		if password == "" {
			c.Next()
			return
		}

		// Always allow login and auth-check endpoints through without a session.
		path := c.Request.URL.Path
		if path == "/api/auth/login" || path == "/api/auth/check" {
			c.Next()
			return
		}

		// Validate the session cookie.
		cookie, err := c.Cookie(SessionCookieName)
		if err != nil || !ValidateSessionToken(cookie, password) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Next()
	}
}

// CreateSessionToken creates an HMAC-SHA256 token using the given password as
// key and a fixed payload. The result is hex-encoded.
func CreateSessionToken(password string) string {
	mac := hmac.New(sha256.New, []byte(password))
	mac.Write([]byte(sessionPayload))
	return hex.EncodeToString(mac.Sum(nil))
}

// ValidateSessionToken checks whether the provided token matches the expected
// HMAC-SHA256 signature for the given password. Comparison is constant-time.
func ValidateSessionToken(token, password string) bool {
	expected := CreateSessionToken(password)
	return hmac.Equal([]byte(token), []byte(expected))
}

// SetSessionCookie creates a session token and sets it as an HTTP-only cookie.
func SetSessionCookie(c *gin.Context, password string) {
	token := CreateSessionToken(password)
	maxAge := int(SessionTTL.Seconds())
	c.SetCookie(SessionCookieName, token, maxAge, "/", "", false, true)
}

// ClearSessionCookie removes the session cookie by setting max age to -1.
func ClearSessionCookie(c *gin.Context) {
	c.SetCookie(SessionCookieName, "", -1, "/", "", false, true)
}
