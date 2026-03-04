package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pi/flash-cards/internal/models"
	"github.com/pi/flash-cards/internal/srs"
)

// StudyHandler provides HTTP handlers for study session endpoints.
type StudyHandler struct {
	db *sql.DB
}

// NewStudyHandler creates a new StudyHandler.
func NewStudyHandler(db *sql.DB) *StudyHandler {
	return &StudyHandler{db: db}
}

// Register mounts the study routes on the given router group.
func (h *StudyHandler) Register(r *gin.RouterGroup) {
	r.GET("/study/next", h.NextCard)
	r.POST("/study/review", h.SubmitReview)
	r.GET("/study/new", h.NewCard)
}

// NextCard returns the next due card for review.
// GET /api/study/next?tag=...
func (h *StudyHandler) NextCard(c *gin.Context) {
	tag := strings.TrimSpace(c.Query("tag"))

	state, card, tags, found, err := h.findDueCard(tag, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch next card"})
		return
	}

	if !found {
		newCount, err := h.countNewCards(tag)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count new cards"})
			return
		}
		c.JSON(http.StatusOK, models.StudyDoneResponse{
			Done:         true,
			NewAvailable: newCount,
		})
		return
	}

	card.Tags = tags
	h.respondWithStudyCard(c, card, state)
}

// NewCard returns the next new (never-reviewed) card.
// GET /api/study/new?tag=...
func (h *StudyHandler) NewCard(c *gin.Context) {
	tag := strings.TrimSpace(c.Query("tag"))

	state, card, tags, found, err := h.findDueCard(tag, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch new card"})
		return
	}

	if !found {
		c.JSON(http.StatusOK, models.StudyDoneResponse{
			Done:         true,
			NewAvailable: 0,
		})
		return
	}

	card.Tags = tags
	h.respondWithStudyCard(c, card, state)
}

// SubmitReview processes a review rating for an SRS state.
// POST /api/study/review
func (h *StudyHandler) SubmitReview(c *gin.Context) {
	var req models.ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch the current SRS state.
	var state models.SRSState
	err := h.db.QueryRow(`
		SELECT id, card_id, direction, ease_factor, interval_days,
		       repetitions, next_review, status, learning_step
		FROM srs_state
		WHERE id = ?
	`, req.SRSStateID).Scan(
		&state.ID, &state.CardID, &state.Direction,
		&state.EaseFactor, &state.IntervalDays,
		&state.Repetitions, &state.NextReview, &state.Status, &state.LearningStep,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "SRS state not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch SRS state"})
		return
	}

	// Snapshot before values for the review event.
	intervalBefore := state.IntervalDays
	easeBefore := state.EaseFactor

	// Compute the new state.
	newState := srs.ProcessReview(&state, req.Rating)

	// Update the SRS state row.
	_, err = h.db.Exec(`
		UPDATE srs_state
		SET ease_factor = ?, interval_days = ?, repetitions = ?,
		    next_review = ?, status = ?, learning_step = ?
		WHERE id = ?
	`,
		newState.EaseFactor, newState.IntervalDays, newState.Repetitions,
		newState.NextReview, newState.Status, newState.LearningStep,
		newState.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update SRS state"})
		return
	}

	// Record the review event.
	_, err = h.db.Exec(`
		INSERT INTO review_events
			(srs_state_id, card_id, direction, rating, reviewed_at,
			 interval_before, interval_after, ease_before, ease_after)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		newState.ID, newState.CardID, newState.Direction, req.Rating, time.Now(),
		intervalBefore, newState.IntervalDays, easeBefore, newState.EaseFactor,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record review event"})
		return
	}

	// Compute a human-readable next interval.
	nextInterval := srs.FormatInterval(&state, req.Rating)

	c.JSON(http.StatusOK, models.ReviewResponse{
		SRSState:     newState,
		NextInterval: nextInterval,
	})
}

// findDueCard queries for the next card to study.
// If newOnly is true, it returns only cards with status='new' and repetitions=0.
// Otherwise, it returns cards where next_review <= now.
func (h *StudyHandler) findDueCard(tag string, newOnly bool) (models.SRSState, models.Card, []string, bool, error) {
	var state models.SRSState
	var card models.Card

	args := make([]interface{}, 0, 4)
	var conditions []string
	joinClause := ""

	// Base conditions: card must not be soft-deleted or suspended.
	conditions = append(conditions, "c.deleted_at IS NULL")
	conditions = append(conditions, "c.suspended = FALSE")

	if newOnly {
		conditions = append(conditions, "s.status = 'new'")
		conditions = append(conditions, "s.repetitions = 0")
	} else {
		conditions = append(conditions, "s.next_review <= ?")
		args = append(args, time.Now())
	}

	if tag != "" {
		joinClause = "JOIN card_tags ct ON ct.card_id = c.id"
		conditions = append(conditions, "ct.tag = ?")
		args = append(args, tag)
	}

	query := fmt.Sprintf(`
		SELECT s.id, s.card_id, s.direction, s.ease_factor, s.interval_days,
		       s.repetitions, s.next_review, s.status, s.learning_step,
		       c.id, c.czech, c.english, c.deleted_at, c.suspended, c.created_at, c.updated_at
		FROM srs_state s
		JOIN cards c ON c.id = s.card_id
		%s
		WHERE %s
		ORDER BY s.next_review ASC
		LIMIT 1
	`, joinClause, strings.Join(conditions, " AND "))

	err := h.db.QueryRow(query, args...).Scan(
		&state.ID, &state.CardID, &state.Direction,
		&state.EaseFactor, &state.IntervalDays,
		&state.Repetitions, &state.NextReview, &state.Status, &state.LearningStep,
		&card.ID, &card.Czech, &card.English, &card.DeletedAt, &card.Suspended,
		&card.CreatedAt, &card.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return state, card, nil, false, nil
	}
	if err != nil {
		return state, card, nil, false, fmt.Errorf("querying due card: %w", err)
	}

	// Fetch tags for this card.
	tags, err := h.fetchCardTags(card.ID)
	if err != nil {
		return state, card, nil, false, fmt.Errorf("fetching card tags: %w", err)
	}

	return state, card, tags, true, nil
}

// countNewCards counts cards with SRS status='new' and repetitions=0,
// optionally filtered by tag.
func (h *StudyHandler) countNewCards(tag string) (int, error) {
	args := make([]interface{}, 0, 2)
	var conditions []string
	joinClause := ""

	conditions = append(conditions, "c.deleted_at IS NULL")
	conditions = append(conditions, "c.suspended = FALSE")
	conditions = append(conditions, "s.status = 'new'")
	conditions = append(conditions, "s.repetitions = 0")

	if tag != "" {
		joinClause = "JOIN card_tags ct ON ct.card_id = c.id"
		conditions = append(conditions, "ct.tag = ?")
		args = append(args, tag)
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM srs_state s
		JOIN cards c ON c.id = s.card_id
		%s
		WHERE %s
	`, joinClause, strings.Join(conditions, " AND "))

	var count int
	err := h.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting new cards: %w", err)
	}
	return count, nil
}

// fetchCardTags retrieves all tags for a given card ID.
func (h *StudyHandler) fetchCardTags(cardID int64) ([]string, error) {
	rows, err := h.db.Query("SELECT tag FROM card_tags WHERE card_id = ? ORDER BY tag", cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// respondWithStudyCard builds a study card response with interval hints.
func (h *StudyHandler) respondWithStudyCard(c *gin.Context, card models.Card, state models.SRSState) {
	hints := make(map[string]string, 4)
	for rating := 1; rating <= 4; rating++ {
		hints[strconv.Itoa(rating)] = srs.FormatInterval(&state, rating)
	}

	c.JSON(http.StatusOK, gin.H{
		"card":          card,
		"srsState":      state,
		"intervalHints": hints,
	})
}
