package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	flashcards "github.com/pi/flash-cards"
	"github.com/pi/flash-cards/internal/auth"
	"github.com/pi/flash-cards/internal/db"
	"github.com/pi/flash-cards/internal/handlers"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	password := os.Getenv("APP_PASSWORD")

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = db.DefaultDBPath
	}

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// API routes
	api := r.Group("/api")
	api.Use(auth.NewMiddleware(password))

	handlers.NewAuthHandler(password).Register(api)
	handlers.NewCardHandler(database).Register(api)
	handlers.NewStudyHandler(database).Register(api)
	handlers.NewImportHandler(database).Register(api)
	handlers.NewStatsHandler(database).Register(api)
	handlers.NewSettingsHandler(database).Register(api)

	// Serve embedded frontend
	distFS, err := fs.Sub(flashcards.FrontendDist, "frontend/dist")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}

	fileServer := http.FileServer(http.FS(distFS))
	r.NoRoute(func(c *gin.Context) {
		// Try to serve the file directly
		path := c.Request.URL.Path
		f, err := distFS.Open(path[1:]) // strip leading /
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		// SPA fallback: serve index.html for any unmatched route
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	log.Printf("Starting server on :%s", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
