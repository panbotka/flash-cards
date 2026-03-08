package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pi/flash-cards/internal/db"
	"github.com/pi/flash-cards/internal/handlers"
	"github.com/pi/flash-cards/internal/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupTestRouter creates a Gin engine backed by an in-memory SQLite database
// with all handler groups registered. Each call returns a fresh database.
func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	dsn := "file::memory:?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on"
	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatalf("opening test database: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	if err := db.RunMigrations(sqlDB); err != nil {
		t.Fatalf("running migrations: %v", err)
	}

	r := gin.New()
	api := r.Group("/api")

	handlers.NewAuthHandler("").Register(api)
	handlers.NewCardHandler(sqlDB).Register(api)
	handlers.NewStudyHandler(sqlDB).Register(api)
	handlers.NewImportHandler(sqlDB).Register(api)
	handlers.NewStatsHandler(sqlDB).Register(api)
	handlers.NewSettingsHandler(sqlDB).Register(api)

	return r
}

// createTestCard creates a card via the API and returns the parsed response.
func createTestCard(t *testing.T, router *gin.Engine, czech, english string, tags []string) models.Card {
	t.Helper()

	body, _ := json.Marshal(models.CreateCardRequest{
		Czech:   czech,
		English: english,
		Tags:    tags,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/cards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("createTestCard: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var card models.Card
	if err := json.Unmarshal(w.Body.Bytes(), &card); err != nil {
		t.Fatalf("createTestCard: failed to parse response: %v", err)
	}
	return card
}
