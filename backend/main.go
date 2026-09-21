package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"time-sync/backend/internal/db"
	"time-sync/backend/internal/handlers"
	"time-sync/backend/internal/middleware"
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
	// Only the reverse proxy (Caddy, on the private compose network) ever
	// connects directly to this process, so its X-Forwarded-For can be
	// trusted to carry the real client IP for rate limiting.
	if err := r.SetTrustedProxies([]string{"172.16.0.0/12"}); err != nil {
		log.Fatalf("failed to set trusted proxies: %v", err)
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{allowedOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	limiter := middleware.NewRateLimiter(rate.Every(2*time.Second), 15)

	api := r.Group("/api")
	api.Use(limiter.Middleware())
	{
		api.GET("/stats", h.GetStats)
		api.POST("/events", h.CreateEvent)
		api.GET("/events/:id", h.GetEvent)
		api.POST("/events/:id/participants", h.AddParticipant)
		api.PUT("/events/:id/participants/:participantId", h.UpdateParticipant)
		api.PUT("/events/:id/finalize", h.FinalizeEvent)
	}

	log.Printf("listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
