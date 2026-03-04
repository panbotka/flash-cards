package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pi/flash-cards/internal/auth"
	"github.com/pi/flash-cards/internal/models"
)

// AuthHandler provides login, logout, and auth-check endpoints.
type AuthHandler struct {
	password string
}

// NewAuthHandler creates a new AuthHandler with the configured password.
func NewAuthHandler(password string) *AuthHandler {
	return &AuthHandler{password: password}
}

// Register mounts auth routes on the provided router group.
func (h *AuthHandler) Register(r *gin.RouterGroup) {
	r.POST("/auth/login", h.Login)
	r.POST("/auth/logout", h.Logout)
	r.GET("/auth/check", h.Check)
}

// Login handles POST /api/auth/login.
// It validates the submitted password and sets a session cookie on success.
func (h *AuthHandler) Login(c *gin.Context) {
	// No auth mode: always succeed.
	if h.password == "" {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Password != h.password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	auth.SetSessionCookie(c, h.password)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Logout handles POST /api/auth/logout.
// It clears the session cookie.
func (h *AuthHandler) Logout(c *gin.Context) {
	auth.ClearSessionCookie(c)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Check handles GET /api/auth/check.
// It reports whether the user is authenticated and whether auth is required.
// This endpoint is always accessible (the middleware skips it).
func (h *AuthHandler) Check(c *gin.Context) {
	authRequired := h.password != ""

	// If no auth is required, the user is always considered authenticated.
	if !authRequired {
		c.JSON(http.StatusOK, gin.H{
			"authenticated": true,
			"authRequired":  false,
		})
		return
	}

	// Auth is required -- check for a valid session cookie.
	cookie, err := c.Cookie(auth.SessionCookieName)
	authenticated := err == nil && auth.ValidateSessionToken(cookie, h.password)

	c.JSON(http.StatusOK, gin.H{
		"authenticated": authenticated,
		"authRequired":  true,
	})
}
