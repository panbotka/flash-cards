package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pi/flash-cards/internal/models"
)

// CardHandler provides CRUD endpoints for flash cards.
type CardHandler struct {
	db *sql.DB
}

// NewCardHandler creates a new CardHandler backed by the given database.
func NewCardHandler(db *sql.DB) *CardHandler {
	return &CardHandler{db: db}
}

// Register mounts card routes on the provided router group.
func (h *CardHandler) Register(r *gin.RouterGroup) {
	r.GET("/tags", h.ListTags)
	r.GET("/cards", h.ListCards)
	r.GET("/cards/:id", h.GetCard)
	r.POST("/cards", h.CreateCard)
	r.PUT("/cards/:id", h.UpdateCard)
	r.DELETE("/cards/:id", h.DeleteCard)
	r.POST("/cards/:id/suspend", h.ToggleSuspend)
	r.POST("/cards/:id/restore", h.RestoreCard)
}

// ListTags returns all distinct tags from non-deleted cards.
func (h *CardHandler) ListTags(c *gin.Context) {
	rows, err := h.db.Query(`SELECT DISTINCT tag FROM card_tags WHERE card_id IN (SELECT id FROM cards WHERE deleted_at IS NULL) ORDER BY tag`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query tags"})
		return
	}
	defer rows.Close()

	tags := []string{}
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan tag"})
			return
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to iterate tags"})
		return
	}

	c.JSON(http.StatusOK, tags)
}

// ListCards returns all non-deleted cards, optionally filtered by tag or search text.
func (h *CardHandler) ListCards(c *gin.Context) {
	tag := strings.TrimSpace(c.Query("tag"))
	search := strings.TrimSpace(c.Query("search"))

	query := `SELECT id, czech, english, deleted_at, suspended, created_at, updated_at FROM cards WHERE deleted_at IS NULL`
	args := []interface{}{}

	if tag != "" {
		query += ` AND id IN (SELECT card_id FROM card_tags WHERE tag = ?)`
		args = append(args, tag)
	}
	if search != "" {
		query += ` AND (czech LIKE ? OR english LIKE ?)`
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query cards"})
		return
	}
	defer rows.Close()

	cards := []models.Card{}
	for rows.Next() {
		var card models.Card
		if err := rows.Scan(&card.ID, &card.Czech, &card.English, &card.DeletedAt, &card.Suspended, &card.CreatedAt, &card.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan card"})
			return
		}
		cards = append(cards, card)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to iterate cards"})
		return
	}

	// Fetch tags for all cards in a single query.
	if len(cards) > 0 {
		ids := make([]string, len(cards))
		for i, card := range cards {
			ids[i] = strconv.FormatInt(card.ID, 10)
		}
		tagQuery := `SELECT card_id, tag FROM card_tags WHERE card_id IN (` + strings.Join(ids, ",") + `) ORDER BY tag`
		tagRows, err := h.db.Query(tagQuery)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query tags"})
			return
		}
		defer tagRows.Close()

		tagMap := map[int64][]string{}
		for tagRows.Next() {
			var cardID int64
			var t string
			if err := tagRows.Scan(&cardID, &t); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan tag"})
				return
			}
			tagMap[cardID] = append(tagMap[cardID], t)
		}
		if err := tagRows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to iterate tags"})
			return
		}

		for i := range cards {
			cards[i].Tags = tagMap[cards[i].ID]
			if cards[i].Tags == nil {
				cards[i].Tags = []string{}
			}
		}
	}

	c.JSON(http.StatusOK, cards)
}

// GetCard returns a single card by ID, including tags and SRS states.
// Includes soft-deleted cards so they can be restored.
func (h *CardHandler) GetCard(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}

	var card models.Card
	err = h.db.QueryRow(
		`SELECT id, czech, english, deleted_at, suspended, created_at, updated_at FROM cards WHERE id = ?`, id,
	).Scan(&card.ID, &card.Czech, &card.English, &card.DeletedAt, &card.Suspended, &card.CreatedAt, &card.UpdatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query card"})
		return
	}

	// Fetch tags.
	card.Tags, err = h.fetchTags(card.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query tags"})
		return
	}

	// Fetch SRS states.
	card.SRSStates, err = h.fetchSRSStates(card.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query srs states"})
		return
	}

	c.JSON(http.StatusOK, card)
}

// CreateCard inserts a new card with tags and initial SRS states for both directions.
func (h *CardHandler) CreateCard(c *gin.Context) {
	var req models.CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Czech = strings.TrimSpace(req.Czech)
	req.English = strings.TrimSpace(req.English)
	if req.Czech == "" || req.English == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "czech and english are required"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	now := time.Now().UTC()

	result, err := tx.Exec(
		`INSERT INTO cards (czech, english, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		req.Czech, req.English, now, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert card"})
		return
	}

	cardID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get card id"})
		return
	}

	// Insert tags.
	for _, tag := range req.Tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO card_tags (card_id, tag) VALUES (?, ?)`, cardID, tag); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert tag"})
			return
		}
	}

	// Create SRS states for both directions.
	for _, dir := range []string{"cz_en", "en_cz"} {
		if _, err := tx.Exec(
			`INSERT INTO srs_state (card_id, direction, ease_factor, interval_days, repetitions, next_review, status, learning_step) VALUES (?, ?, 2.5, 0, 0, ?, 'new', 0)`,
			cardID, dir, now,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert srs state"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	// Read back the full card.
	var card models.Card
	err = h.db.QueryRow(
		`SELECT id, czech, english, deleted_at, suspended, created_at, updated_at FROM cards WHERE id = ?`, cardID,
	).Scan(&card.ID, &card.Czech, &card.English, &card.DeletedAt, &card.Suspended, &card.CreatedAt, &card.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read back card"})
		return
	}

	card.Tags, _ = h.fetchTags(card.ID)
	card.SRSStates, _ = h.fetchSRSStates(card.ID)

	c.JSON(http.StatusCreated, card)
}

// UpdateCard performs a partial update on an existing card.
func (h *CardHandler) UpdateCard(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}

	var req models.UpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify card exists and is not deleted.
	var exists bool
	err = h.db.QueryRow(`SELECT 1 FROM cards WHERE id = ? AND deleted_at IS NULL`, id).Scan(&exists)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query card"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	now := time.Now().UTC()

	// Build dynamic UPDATE for non-nil fields.
	setClauses := []string{"updated_at = ?"}
	args := []interface{}{now}

	if req.Czech != nil {
		v := strings.TrimSpace(*req.Czech)
		if v == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "czech cannot be empty"})
			return
		}
		setClauses = append(setClauses, "czech = ?")
		args = append(args, v)
	}
	if req.English != nil {
		v := strings.TrimSpace(*req.English)
		if v == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "english cannot be empty"})
			return
		}
		setClauses = append(setClauses, "english = ?")
		args = append(args, v)
	}

	args = append(args, id)
	updateQuery := `UPDATE cards SET ` + strings.Join(setClauses, ", ") + ` WHERE id = ?`

	if _, err := tx.Exec(updateQuery, args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update card"})
		return
	}

	// If tags are provided, replace them.
	if req.Tags != nil {
		if _, err := tx.Exec(`DELETE FROM card_tags WHERE card_id = ?`, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete old tags"})
			return
		}
		for _, tag := range *req.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			if _, err := tx.Exec(`INSERT INTO card_tags (card_id, tag) VALUES (?, ?)`, id, tag); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert tag"})
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	// Read back the updated card.
	var card models.Card
	err = h.db.QueryRow(
		`SELECT id, czech, english, deleted_at, suspended, created_at, updated_at FROM cards WHERE id = ?`, id,
	).Scan(&card.ID, &card.Czech, &card.English, &card.DeletedAt, &card.Suspended, &card.CreatedAt, &card.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read back card"})
		return
	}

	card.Tags, _ = h.fetchTags(card.ID)
	card.SRSStates, _ = h.fetchSRSStates(card.ID)

	c.JSON(http.StatusOK, card)
}

// DeleteCard performs a soft delete by setting deleted_at.
func (h *CardHandler) DeleteCard(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}

	result, err := h.db.Exec(`UPDATE cards SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL`, time.Now().UTC(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete card"})
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check affected rows"})
		return
	}
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ToggleSuspend toggles the suspended flag on a card.
func (h *CardHandler) ToggleSuspend(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}

	result, err := h.db.Exec(
		`UPDATE cards SET suspended = NOT suspended, updated_at = ? WHERE id = ? AND deleted_at IS NULL`, time.Now().UTC(), id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to toggle suspend"})
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check affected rows"})
		return
	}
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
		return
	}

	var card models.Card
	err = h.db.QueryRow(
		`SELECT id, czech, english, deleted_at, suspended, created_at, updated_at FROM cards WHERE id = ?`, id,
	).Scan(&card.ID, &card.Czech, &card.English, &card.DeletedAt, &card.Suspended, &card.CreatedAt, &card.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read back card"})
		return
	}

	card.Tags, _ = h.fetchTags(card.ID)
	card.SRSStates, _ = h.fetchSRSStates(card.ID)

	c.JSON(http.StatusOK, card)
}

// RestoreCard clears the deleted_at timestamp to restore a soft-deleted card.
func (h *CardHandler) RestoreCard(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card id"})
		return
	}

	result, err := h.db.Exec(
		`UPDATE cards SET deleted_at = NULL, updated_at = ? WHERE id = ? AND deleted_at IS NOT NULL`, time.Now().UTC(), id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restore card"})
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check affected rows"})
		return
	}
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "card not found or not deleted"})
		return
	}

	var card models.Card
	err = h.db.QueryRow(
		`SELECT id, czech, english, deleted_at, suspended, created_at, updated_at FROM cards WHERE id = ?`, id,
	).Scan(&card.ID, &card.Czech, &card.English, &card.DeletedAt, &card.Suspended, &card.CreatedAt, &card.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read back card"})
		return
	}

	card.Tags, _ = h.fetchTags(card.ID)
	card.SRSStates, _ = h.fetchSRSStates(card.ID)

	c.JSON(http.StatusOK, card)
}

// fetchTags retrieves all tags for a given card ID.
func (h *CardHandler) fetchTags(cardID int64) ([]string, error) {
	rows, err := h.db.Query(`SELECT tag FROM card_tags WHERE card_id = ? ORDER BY tag`, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []string{}
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// fetchSRSStates retrieves SRS states for a given card ID.
func (h *CardHandler) fetchSRSStates(cardID int64) ([]models.SRSState, error) {
	rows, err := h.db.Query(
		`SELECT id, card_id, direction, ease_factor, interval_days, repetitions, next_review, status, learning_step FROM srs_state WHERE card_id = ? ORDER BY direction`,
		cardID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := []models.SRSState{}
	for rows.Next() {
		var s models.SRSState
		if err := rows.Scan(&s.ID, &s.CardID, &s.Direction, &s.EaseFactor, &s.IntervalDays, &s.Repetitions, &s.NextReview, &s.Status, &s.LearningStep); err != nil {
			return nil, err
		}
		states = append(states, s)
	}
	return states, rows.Err()
}
