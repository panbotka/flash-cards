package handlers

import (
	"database/sql"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// StatsHandler provides endpoints for study statistics and analytics.
type StatsHandler struct {
	db *sql.DB
}

// NewStatsHandler creates a new StatsHandler backed by the given database.
func NewStatsHandler(db *sql.DB) *StatsHandler {
	return &StatsHandler{db: db}
}

// Register mounts stats routes on the provided router group.
func (h *StatsHandler) Register(r *gin.RouterGroup) {
	r.GET("/stats/summary", h.Summary)
	r.GET("/stats/heatmap", h.Heatmap)
	r.GET("/stats/accuracy", h.Accuracy)
	r.GET("/stats/maturity", h.Maturity)
	r.GET("/stats/forecast", h.Forecast)
	r.GET("/stats/hardest", h.Hardest)
}

// Summary returns high-level review statistics.
// GET /api/stats/summary
func (h *StatsHandler) Summary(c *gin.Context) {
	today := time.Now().Format("2006-01-02")

	var reviewsToday int
	err := h.db.QueryRow(
		`SELECT COUNT(*) FROM review_events WHERE date(reviewed_at) = ?`, today,
	).Scan(&reviewsToday)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count today's reviews"})
		return
	}

	var totalCards int
	err = h.db.QueryRow(
		`SELECT COUNT(*) FROM cards WHERE deleted_at IS NULL`,
	).Scan(&totalCards)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count cards"})
		return
	}

	// Calculate streak: count consecutive days going back from today that have reviews.
	streak, err := h.calculateStreak(today)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate streak"})
		return
	}

	// Calculate today's accuracy: fraction of reviews with rating >= 3.
	var accuracyToday float64
	if reviewsToday > 0 {
		var goodCount int
		err = h.db.QueryRow(
			`SELECT COUNT(*) FROM review_events WHERE date(reviewed_at) = ? AND rating >= 3`, today,
		).Scan(&goodCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate accuracy"})
			return
		}
		accuracyToday = math.Round(float64(goodCount)/float64(reviewsToday)*100) / 100
	}

	c.JSON(http.StatusOK, gin.H{
		"reviewsToday": reviewsToday,
		"totalCards":   totalCards,
		"streak":       streak,
		"accuracyToday": accuracyToday,
	})
}

// calculateStreak counts consecutive days (going back from today) with at least one review.
func (h *StatsHandler) calculateStreak(today string) (int, error) {
	rows, err := h.db.Query(
		`SELECT DISTINCT date(reviewed_at) AS d FROM review_events ORDER BY d DESC`,
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	streak := 0
	expected, err := time.Parse("2006-01-02", today)
	if err != nil {
		return 0, err
	}

	for rows.Next() {
		var dateStr string
		if err := rows.Scan(&dateStr); err != nil {
			return 0, err
		}
		reviewDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return 0, err
		}

		if reviewDate.Equal(expected) {
			streak++
			expected = expected.AddDate(0, 0, -1)
		} else if reviewDate.Before(expected) {
			// Gap found; stop counting.
			break
		}
		// If reviewDate is after expected (shouldn't happen with DESC order), skip.
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	return streak, nil
}

// Heatmap returns daily review counts for the last 365 days.
// GET /api/stats/heatmap
func (h *StatsHandler) Heatmap(c *gin.Context) {
	now := time.Now()
	startDate := now.AddDate(0, 0, -364) // 365 days including today

	rows, err := h.db.Query(
		`SELECT date(reviewed_at) AS d, COUNT(*) AS cnt
		 FROM review_events
		 WHERE date(reviewed_at) >= date(?)
		 GROUP BY d
		 ORDER BY d`,
		startDate.Format("2006-01-02"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query heatmap data"})
		return
	}
	defer rows.Close()

	// Build a map of date -> count from the query results.
	countMap := make(map[string]int)
	for rows.Next() {
		var d string
		var cnt int
		if err := rows.Scan(&d, &cnt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan heatmap row"})
			return
		}
		countMap[d] = cnt
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to iterate heatmap rows"})
		return
	}

	// Build the full 365-day array, filling zeros for missing days.
	type heatmapEntry struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}

	result := make([]heatmapEntry, 0, 365)
	for i := 0; i < 365; i++ {
		d := startDate.AddDate(0, 0, i).Format("2006-01-02")
		cnt := countMap[d]
		result = append(result, heatmapEntry{Date: d, Count: cnt})
	}

	c.JSON(http.StatusOK, result)
}

// Accuracy returns daily accuracy for the last 30 days.
// GET /api/stats/accuracy
func (h *StatsHandler) Accuracy(c *gin.Context) {
	startDate := time.Now().AddDate(0, 0, -29).Format("2006-01-02")

	rows, err := h.db.Query(
		`SELECT date(reviewed_at) AS d,
		        COUNT(*) AS total,
		        SUM(CASE WHEN rating >= 3 THEN 1 ELSE 0 END) AS good
		 FROM review_events
		 WHERE date(reviewed_at) >= date(?)
		 GROUP BY d
		 ORDER BY d`,
		startDate,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query accuracy data"})
		return
	}
	defer rows.Close()

	type accuracyEntry struct {
		Date     string  `json:"date"`
		Accuracy float64 `json:"accuracy"`
		Total    int     `json:"total"`
	}

	result := []accuracyEntry{}
	for rows.Next() {
		var d string
		var total, good int
		if err := rows.Scan(&d, &total, &good); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan accuracy row"})
			return
		}
		acc := math.Round(float64(good)/float64(total)*100) / 100
		result = append(result, accuracyEntry{Date: d, Accuracy: acc, Total: total})
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to iterate accuracy rows"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Maturity returns the distribution of cards by SRS maturity level.
// GET /api/stats/maturity
//
// Maturity buckets: new < learning < young (review, interval < 21) < mature (review, interval >= 21).
func (h *StatsHandler) Maturity(c *gin.Context) {
	// For each non-deleted, non-suspended card, find the worst maturity bucket
	// across its directions. We assign a numeric rank:
	//   new=0, learning=1, young=2, mature=3
	// and take the MIN per card_id.
	rows, err := h.db.Query(`
		SELECT
			MIN(CASE
				WHEN s.status = 'new' THEN 0
				WHEN s.status = 'learning' THEN 1
				WHEN s.status = 'review' AND s.interval_days < 21 THEN 2
				ELSE 3
			END) AS maturity_rank
		FROM srs_state s
		JOIN cards c ON c.id = s.card_id
		WHERE c.deleted_at IS NULL AND c.suspended = FALSE
		GROUP BY s.card_id
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query maturity data"})
		return
	}
	defer rows.Close()

	counts := map[string]int{
		"new":      0,
		"learning": 0,
		"young":    0,
		"mature":   0,
	}

	for rows.Next() {
		var rank int
		if err := rows.Scan(&rank); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan maturity row"})
			return
		}
		switch rank {
		case 0:
			counts["new"]++
		case 1:
			counts["learning"]++
		case 2:
			counts["young"]++
		case 3:
			counts["mature"]++
		}
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to iterate maturity rows"})
		return
	}

	c.JSON(http.StatusOK, counts)
}

// Forecast returns the count of upcoming due reviews per day for the next 30 days.
// GET /api/stats/forecast
func (h *StatsHandler) Forecast(c *gin.Context) {
	now := time.Now()
	today := now.Format("2006-01-02")

	rows, err := h.db.Query(
		`SELECT date(s.next_review) AS d, COUNT(*) AS cnt
		 FROM srs_state s
		 JOIN cards c ON c.id = s.card_id
		 WHERE c.deleted_at IS NULL
		   AND c.suspended = FALSE
		   AND date(s.next_review) >= date(?)
		   AND date(s.next_review) < date(?, '+30 days')
		 GROUP BY d
		 ORDER BY d`,
		today, today,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query forecast data"})
		return
	}
	defer rows.Close()

	countMap := make(map[string]int)
	for rows.Next() {
		var d string
		var cnt int
		if err := rows.Scan(&d, &cnt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan forecast row"})
			return
		}
		countMap[d] = cnt
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to iterate forecast rows"})
		return
	}

	type forecastEntry struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}

	todayTime, _ := time.Parse("2006-01-02", today)
	result := make([]forecastEntry, 0, 30)
	for i := 0; i < 30; i++ {
		d := todayTime.AddDate(0, 0, i).Format("2006-01-02")
		cnt := countMap[d]
		result = append(result, forecastEntry{Date: d, Count: cnt})
	}

	c.JSON(http.StatusOK, result)
}

// Hardest returns the 20 cards with the lowest review accuracy.
// GET /api/stats/hardest
func (h *StatsHandler) Hardest(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT
			r.card_id,
			c.czech,
			c.english,
			COUNT(*) AS total_reviews,
			SUM(CASE WHEN r.rating = 1 THEN 1 ELSE 0 END) AS again_count,
			CAST(SUM(CASE WHEN r.rating >= 3 THEN 1 ELSE 0 END) AS REAL) / COUNT(*) AS accuracy
		FROM review_events r
		JOIN cards c ON c.id = r.card_id
		WHERE c.deleted_at IS NULL
		GROUP BY r.card_id
		HAVING COUNT(*) >= 3
		ORDER BY accuracy ASC
		LIMIT 20
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query hardest cards"})
		return
	}
	defer rows.Close()

	type hardestEntry struct {
		CardID       int64   `json:"cardId"`
		Czech        string  `json:"czech"`
		English      string  `json:"english"`
		TotalReviews int     `json:"totalReviews"`
		AgainCount   int     `json:"againCount"`
		Accuracy     float64 `json:"accuracy"`
	}

	result := []hardestEntry{}
	for rows.Next() {
		var e hardestEntry
		if err := rows.Scan(&e.CardID, &e.Czech, &e.English, &e.TotalReviews, &e.AgainCount, &e.Accuracy); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan hardest card row"})
			return
		}
		e.Accuracy = math.Round(e.Accuracy*100) / 100
		result = append(result, e)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to iterate hardest card rows"})
		return
	}

	c.JSON(http.StatusOK, result)
}
