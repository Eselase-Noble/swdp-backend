package database

import (
	"context"
	"log"
	"web-based-dev-platform-backend/internal/config"
	"web-based-dev-platform-backend/internal/users"

	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(databaseUrl string) *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	return pool
}

func ConnectDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	//Auto Migrate the user entity
	db.AutoMigrate(
		&users.User{},
	)

	return db
}
