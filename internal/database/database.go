package database

import (
	"context"
	"log"
	"web-based-dev-platform-backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect returns a pgx connection pool used by the auth layer.
func Connect(databaseUrl string) *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	return pool
}

// ConnectDB returns a GORM instance used by all domain repositories.
// Schema is managed by SQL migrations in /migrations — no AutoMigrate.
func ConnectDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open GORM connection: %v", err)
	}
	return db
}
