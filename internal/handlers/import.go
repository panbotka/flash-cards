package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pi/flash-cards/internal/importer"
	"github.com/pi/flash-cards/internal/models"
)

// ImportHandler handles bulk card import with preview and commit steps.
type ImportHandler struct {
	db *sql.DB
}

// NewImportHandler creates a new ImportHandler.
func NewImportHandler(db *sql.DB) *ImportHandler {
	return &ImportHandler{db: db}
}

// Register mounts the import routes on the given router group.
func (h *ImportHandler) Register(r *gin.RouterGroup) {
	r.POST("/cards/import/preview", h.Preview)
	r.POST("/cards/import/commit", h.Commit)
}

// previewCard is a single card in the preview response, annotated with
// whether it already exists in the database.
type previewCard struct {
	Czech       string `json:"czech"`
	English     string `json:"english"`
	IsDuplicate bool   `json:"isDuplicate"`
}

// previewResponse is the JSON payload returned by Preview.
type previewResponse struct {
	Cards      []previewCard `json:"cards"`
	Duplicates int           `json:"duplicates"`
	Total      int           `json:"total"`
}

// commitResponse is the JSON payload returned by Commit.
type commitResponse struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
}

// Preview parses the submitted content, detects duplicates against the
// existing card database, and returns annotated results without persisting
// anything.
func (h *ImportHandler) Preview(c *gin.Context) {
	var req models.ImportPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	cards, err := importer.Parse(req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(cards) == 0 {
		c.JSON(http.StatusOK, previewResponse{
			Cards:      []previewCard{},
			Duplicates: 0,
			Total:      0,
		})
		return
	}

	duplicates := 0
	result := make([]previewCard, 0, len(cards))

	for _, card := range cards {
		dup, err := h.isDuplicate(card.Czech, card.English)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		if dup {
			duplicates++
		}
		result = append(result, previewCard{
			Czech:       card.Czech,
			English:     card.English,
			IsDuplicate: dup,
		})
	}

	c.JSON(http.StatusOK, previewResponse{
		Cards:      result,
		Duplicates: duplicates,
		Total:      len(result),
	})
}

// Commit inserts non-duplicate cards into the database within a single
// transaction, creating associated tags and SRS state entries.
func (h *ImportHandler) Commit(c *gin.Context) {
	var req models.ImportCommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cards are required"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	imported := 0
	skipped := 0
	now := time.Now().UTC()

	for _, card := range req.Cards {
		czech := strings.TrimSpace(card.Czech)
		english := strings.TrimSpace(card.English)
		if czech == "" || english == "" {
			skipped++
			continue
		}

		dup, err := isDuplicateTx(tx, czech, english)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		if dup {
			skipped++
			continue
		}

		// Insert the card.
		res, err := tx.Exec(
			`INSERT INTO cards (czech, english, created_at, updated_at) VALUES (?, ?, ?, ?)`,
			czech, english, now, now,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert card"})
			return
		}

		cardID, err := res.LastInsertId()
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
			if _, err := tx.Exec(
				`INSERT INTO card_tags (card_id, tag) VALUES (?, ?)`,
				cardID, tag,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert tag"})
				return
			}
		}

		// Create SRS state entries for both directions.
		for _, direction := range []string{"cz_en", "en_cz"} {
			if _, err := tx.Exec(
				`INSERT INTO srs_state (card_id, direction, ease_factor, interval_days, repetitions, next_review, status, learning_step)
				 VALUES (?, ?, 2.5, 0, 0, ?, 'new', 0)`,
				cardID, direction, now,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create srs state"})
				return
			}
		}

		imported++
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, commitResponse{
		Imported: imported,
		Skipped:  skipped,
	})
}

// isDuplicate checks whether a card with the same czech and english text
// already exists (case-insensitive) and is not soft-deleted.
func (h *ImportHandler) isDuplicate(czech, english string) (bool, error) {
	var count int
	err := h.db.QueryRow(
		`SELECT COUNT(*) FROM cards
		 WHERE LOWER(czech) = LOWER(?) AND LOWER(english) = LOWER(?) AND deleted_at IS NULL`,
		czech, english,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// isDuplicateTx is the same duplicate check but runs within a transaction.
func isDuplicateTx(tx *sql.Tx, czech, english string) (bool, error) {
	var count int
	err := tx.QueryRow(
		`SELECT COUNT(*) FROM cards
		 WHERE LOWER(czech) = LOWER(?) AND LOWER(english) = LOWER(?) AND deleted_at IS NULL`,
		czech, english,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
