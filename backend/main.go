package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"time-sync/backend/internal/db"
	"time-sync/backend/internal/handlers"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	databaseURL := getenv("DATABASE_URL", "postgres://timesync:timesync@localhost:5432/timesync?sslmode=disable")
	port := getenv("PORT", "8080")
	publicURL := getenv("PUBLIC_URL", "http://localhost:5173")
	allowedOrigin := getenv("ALLOWED_ORIGIN", "http://localhost:5173")

	ctx := context.Background()
	pool, err := db.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	h := handlers.New(pool, publicURL)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{allowedOrigin},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		api.POST("/events", h.CreateEvent)
		api.GET("/events/:id", h.GetEvent)
		api.POST("/events/:id/participants", h.AddParticipant)
	}

	log.Printf("listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
