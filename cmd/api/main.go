package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/guruorgoru/carevo/internal/auth"
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

	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	r := router.New(db, jwtService)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("listening on %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server: %v", err)
		os.Exit(1)
	}
}
