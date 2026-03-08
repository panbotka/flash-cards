package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SettingsHandler provides endpoints for app settings.
type SettingsHandler struct {
	db *sql.DB
}

// NewSettingsHandler creates a new SettingsHandler backed by the given database.
func NewSettingsHandler(db *sql.DB) *SettingsHandler {
	return &SettingsHandler{db: db}
}

// Register mounts settings routes on the provided router group.
func (h *SettingsHandler) Register(r *gin.RouterGroup) {
	r.GET("/settings/daily-goal", h.GetDailyGoal)
	r.PUT("/settings/daily-goal", h.SetDailyGoal)
}

// GetDailyGoal returns the current daily review goal.
// GET /api/settings/daily-goal
func (h *SettingsHandler) GetDailyGoal(c *gin.Context) {
	var value string
	err := h.db.QueryRow(`SELECT value FROM settings WHERE key = 'daily_goal'`).Scan(&value)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{"goal": 0})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read daily goal"})
		return
	}

	goal, err := strconv.Atoi(value)
	if err != nil {
		goal = 0
	}
	c.JSON(http.StatusOK, gin.H{"goal": goal})
}

// SetDailyGoal updates the daily review goal.
// PUT /api/settings/daily-goal
func (h *SettingsHandler) SetDailyGoal(c *gin.Context) {
	var req struct {
		Goal int `json:"goal"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Goal < 0 {
		req.Goal = 0
	}

	_, err := h.db.Exec(
		`INSERT INTO settings (key, value) VALUES ('daily_goal', ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		strconv.Itoa(req.Goal),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save daily goal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"goal": req.Goal})
}
