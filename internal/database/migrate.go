package database

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dsn string) error {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		log.Printf("migrations: applied (version %d, dirty %v)", version, dirty)
	} else {
		log.Printf("migrations: up to date (version %d)", version)
	}

	return nil
}
