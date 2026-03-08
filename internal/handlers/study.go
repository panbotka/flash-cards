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
	r.POST("/study/undo", h.UndoReview)
	r.GET("/study/new", h.NewCard)
}

// validDirection checks if a direction value is allowed.
func validDirection(d string) bool {
	return d == "cz_en" || d == "en_cz"
}

// NextCard returns the next due card for review.
// GET /api/study/next?tag=...&direction=...&mode=cram&exclude=1,2,3
func (h *StudyHandler) NextCard(c *gin.Context) {
	tag := strings.TrimSpace(c.Query("tag"))
	direction := c.DefaultQuery("direction", "cz_en")
	if !validDirection(direction) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid direction"})
		return
	}

	mode := c.Query("mode")
	if mode == "cram" {
		// Parse excluded SRS state IDs (already-seen cards in this cram session).
		var excludeIDs []int64
		if exc := c.Query("exclude"); exc != "" {
			for _, s := range strings.Split(exc, ",") {
				if id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
					excludeIDs = append(excludeIDs, id)
				}
			}
		}
		state, card, tags, found, err := h.findCramCard(tag, direction, excludeIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch cram card"})
			return
		}
		if !found {
			c.JSON(http.StatusOK, models.StudyDoneResponse{Done: true, NewAvailable: 0})
			return
		}
		card.Tags = tags
		h.respondWithStudyCard(c, card, state)
		return
	}

	state, card, tags, found, err := h.findDueCard(tag, direction, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch next card"})
		return
	}

	if !found {
		newCount, err := h.countNewCards(tag, direction)
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
// GET /api/study/new?tag=...&direction=...
func (h *StudyHandler) NewCard(c *gin.Context) {
	tag := strings.TrimSpace(c.Query("tag"))
	direction := c.DefaultQuery("direction", "cz_en")
	if !validDirection(direction) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid direction"})
		return
	}

	state, card, tags, found, err := h.findDueCard(tag, direction, true)
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

	if req.Cram {
		// Cram mode: log the review event but do NOT update SRS state.
		_, err = h.db.Exec(`
			INSERT INTO review_events
				(srs_state_id, card_id, direction, rating, reviewed_at, cram)
			VALUES (?, ?, ?, ?, ?, TRUE)
		`, state.ID, state.CardID, state.Direction, req.Rating, time.Now())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record review event"})
			return
		}

		c.JSON(http.StatusOK, models.ReviewResponse{
			SRSState:     state,
			NextInterval: "",
		})
		return
	}

	// Snapshot full before-state for the review event (enables undo).
	beforeState := state

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

	// Record the review event with full before-state for undo support.
	_, err = h.db.Exec(`
		INSERT INTO review_events
			(srs_state_id, card_id, direction, rating, reviewed_at,
			 interval_before, interval_after, ease_before, ease_after,
			 status_before, learning_step_before, repetitions_before, next_review_before)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		newState.ID, newState.CardID, newState.Direction, req.Rating, time.Now(),
		beforeState.IntervalDays, newState.IntervalDays, beforeState.EaseFactor, newState.EaseFactor,
		beforeState.Status, beforeState.LearningStep, beforeState.Repetitions, beforeState.NextReview,
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

// UndoReview reverts the most recent review for the given direction/tag.
// POST /api/study/undo?direction=...&tag=...
func (h *StudyHandler) UndoReview(c *gin.Context) {
	direction := c.DefaultQuery("direction", "cz_en")
	if !validDirection(direction) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid direction"})
		return
	}
	tag := strings.TrimSpace(c.Query("tag"))

	// Find the most recent review event for this direction (and optional tag).
	args := []interface{}{direction}
	tagJoin := ""
	tagWhere := ""
	if tag != "" {
		tagJoin = "JOIN card_tags ct ON ct.card_id = re.card_id"
		tagWhere = "AND ct.tag = ?"
		args = append(args, tag)
	}

	query := fmt.Sprintf(`
		SELECT re.id, re.srs_state_id, re.card_id,
		       re.interval_before, re.ease_before,
		       re.status_before, re.learning_step_before,
		       re.repetitions_before, re.next_review_before
		FROM review_events re
		%s
		WHERE re.direction = ? %s
		ORDER BY re.reviewed_at DESC
		LIMIT 1
	`, tagJoin, tagWhere)

	var (
		eventID          int64
		srsStateID       int64
		cardID           int64
		intervalBefore   *float64
		easeBefore       *float64
		statusBefore     *string
		learningStepBef  *int
		repetitionsBef   *int
		nextReviewBefore *time.Time
	)
	err := h.db.QueryRow(query, args...).Scan(
		&eventID, &srsStateID, &cardID,
		&intervalBefore, &easeBefore,
		&statusBefore, &learningStepBef,
		&repetitionsBef, &nextReviewBefore,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "no review to undo"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find review event"})
		return
	}

	// Require full before-state (old events before migration 4 won't have these).
	if statusBefore == nil || learningStepBef == nil || repetitionsBef == nil || nextReviewBefore == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no review to undo"})
		return
	}

	// Restore the SRS state and delete the review event in a transaction.
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}

	_, err = tx.Exec(`
		UPDATE srs_state
		SET ease_factor = ?, interval_days = ?, repetitions = ?,
		    next_review = ?, status = ?, learning_step = ?
		WHERE id = ?
	`, *easeBefore, *intervalBefore, *repetitionsBef,
		*nextReviewBefore, *statusBefore, *learningStepBef,
		srsStateID,
	)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restore SRS state"})
		return
	}

	_, err = tx.Exec("DELETE FROM review_events WHERE id = ?", eventID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete review event"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit undo"})
		return
	}

	// Return the restored card for re-display.
	var card models.Card
	err = h.db.QueryRow(`
		SELECT id, czech, english, deleted_at, suspended, created_at, updated_at
		FROM cards WHERE id = ?
	`, cardID).Scan(
		&card.ID, &card.Czech, &card.English, &card.DeletedAt, &card.Suspended,
		&card.CreatedAt, &card.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch card"})
		return
	}

	tags, err := h.fetchCardTags(cardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tags"})
		return
	}
	card.Tags = tags

	// Fetch the restored SRS state.
	var state models.SRSState
	err = h.db.QueryRow(`
		SELECT id, card_id, direction, ease_factor, interval_days,
		       repetitions, next_review, status, learning_step
		FROM srs_state WHERE id = ?
	`, srsStateID).Scan(
		&state.ID, &state.CardID, &state.Direction,
		&state.EaseFactor, &state.IntervalDays,
		&state.Repetitions, &state.NextReview, &state.Status, &state.LearningStep,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch restored state"})
		return
	}

	h.respondWithStudyCard(c, card, state)
}

// findDueCard queries for the next card to study.
// If newOnly is true, it returns only cards with status='new' and repetitions=0.
// Otherwise, it returns cards where next_review <= now.
func (h *StudyHandler) findDueCard(tag string, direction string, newOnly bool) (models.SRSState, models.Card, []string, bool, error) {
	var state models.SRSState
	var card models.Card

	args := make([]interface{}, 0, 4)
	var conditions []string
	joinClause := ""

	// Base conditions: card must not be soft-deleted or suspended.
	conditions = append(conditions, "c.deleted_at IS NULL")
	conditions = append(conditions, "c.suspended = FALSE")
	conditions = append(conditions, "s.direction = ?")
	args = append(args, direction)

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

// findCramCard returns a random card for cram mode, excluding already-seen SRS state IDs.
func (h *StudyHandler) findCramCard(tag string, direction string, excludeIDs []int64) (models.SRSState, models.Card, []string, bool, error) {
	var state models.SRSState
	var card models.Card

	args := make([]interface{}, 0, 4)
	var conditions []string
	joinClause := ""

	conditions = append(conditions, "c.deleted_at IS NULL")
	conditions = append(conditions, "c.suspended = FALSE")
	conditions = append(conditions, "s.direction = ?")
	args = append(args, direction)

	if tag != "" {
		joinClause = "JOIN card_tags ct ON ct.card_id = c.id"
		conditions = append(conditions, "ct.tag = ?")
		args = append(args, tag)
	}

	if len(excludeIDs) > 0 {
		placeholders := make([]string, len(excludeIDs))
		for i, id := range excludeIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		conditions = append(conditions, fmt.Sprintf("s.id NOT IN (%s)", strings.Join(placeholders, ",")))
	}

	query := fmt.Sprintf(`
		SELECT s.id, s.card_id, s.direction, s.ease_factor, s.interval_days,
		       s.repetitions, s.next_review, s.status, s.learning_step,
		       c.id, c.czech, c.english, c.deleted_at, c.suspended, c.created_at, c.updated_at
		FROM srs_state s
		JOIN cards c ON c.id = s.card_id
		%s
		WHERE %s
		ORDER BY RANDOM()
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
		return state, card, nil, false, fmt.Errorf("querying cram card: %w", err)
	}

	tags, err := h.fetchCardTags(card.ID)
	if err != nil {
		return state, card, nil, false, fmt.Errorf("fetching card tags: %w", err)
	}

	return state, card, tags, true, nil
}

// countNewCards counts cards with SRS status='new' and repetitions=0,
// optionally filtered by tag.
func (h *StudyHandler) countNewCards(tag string, direction string) (int, error) {
	args := make([]interface{}, 0, 2)
	var conditions []string
	joinClause := ""

	conditions = append(conditions, "c.deleted_at IS NULL")
	conditions = append(conditions, "c.suspended = FALSE")
	conditions = append(conditions, "s.direction = ?")
	args = append(args, direction)
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
