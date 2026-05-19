// @title           Carevo API
// @version         1.0
// @description     Career discovery API for Nepal
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/guruorgoru/carevo/docs"

	"github.com/guruorgoru/carevo/internal/ai"
	"github.com/guruorgoru/carevo/internal/auth"
	"github.com/guruorgoru/carevo/internal/cache"
	"github.com/guruorgoru/carevo/internal/config"
	"github.com/guruorgoru/carevo/internal/database"
	"github.com/guruorgoru/carevo/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	log.Println("running database migrations...")
	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	cacheStore := cache.New(cfg.RedisURL)
	defer cacheStore.Close()

	var aiProvider *ai.GeminiProvider
	if cfg.GeminiAPIKey != "" {
		aiProvider = ai.NewGeminiProvider(cfg.GeminiAPIKey, cfg.GeminiModel)
	}
	aiWorker := ai.NewWorker(db, aiProvider)
	aiWorker.Start(context.Background())

	r := router.New(db, jwtService, cacheStore, aiWorker, aiProvider)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("listening on %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server: %v", err)
		os.Exit(1)
	}
}
