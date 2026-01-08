package database

import (
	"context"
	"embed"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func RunMigrations(db *pgxpool.Pool) {
	ctx := context.Background()

	files, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		sql, err := migrationFiles.ReadFile("migrations/" + file.Name())
		if err != nil {
			log.Fatal(err)
		}

		if _, err := db.Exec(ctx, string(sql)); err != nil {
			log.Fatalf("migration %s failed: %v", file.Name(), err)
		}
	}
}
