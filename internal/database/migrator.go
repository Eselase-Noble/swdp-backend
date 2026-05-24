package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// migration holds a version tag and its SQL content.
// Each SQL block must be idempotent (IF NOT EXISTS / CREATE TYPE ... IF NOT EXISTS).
type migration struct {
	version string
	sql     string
}

var migrations = []migration{
	{"001_init", sqlInit},
	{"002_indexes", sqlIndexes},
}

// Migrate runs every migration that has not yet been recorded in schema_migrations.
// Safe to call on every startup.
func Migrate(ctx context.Context, pool *pgxpool.Pool) {
	// Ensure the tracking table exists.
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     TEXT PRIMARY KEY,
			applied_at  TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		log.Fatalf("migrator: create schema_migrations: %v", err)
	}

	for _, m := range migrations {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`,
			m.version,
		).Scan(&exists)
		if err != nil {
			log.Fatalf("migrator: check version %s: %v", m.version, err)
		}
		if exists {
			continue
		}

		if _, err := pool.Exec(ctx, m.sql); err != nil {
			log.Fatalf("migrator: apply %s: %v", m.version, err)
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, m.version,
		); err != nil {
			log.Fatalf("migrator: record %s: %v", m.version, err)
		}
		log.Printf("migrator: applied %s", m.version)
	}
}

// ─── SQL content ─────────────────────────────────────────────────────────────

const sqlInit = `
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    user_id       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    username      TEXT        UNIQUE NOT NULL,
    name          TEXT        NOT NULL,
    email         TEXT        UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL,
    created_by    UUID,
    updated_by    UUID,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_yn    BOOLEAN     NOT NULL DEFAULT false
);

DO $$ BEGIN
    CREATE TYPE environment_enum AS ENUM ('dev', 'staging', 'prod');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS projects (
    project_id   UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    project_name TEXT             NOT NULL,
    description  TEXT,
    owner_id     UUID             NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    version      INT,
    environment  environment_enum NOT NULL DEFAULT 'dev',
    created_by   UUID,
    updated_by   UUID,
    created_at   TIMESTAMPTZ      NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ      NOT NULL DEFAULT now(),
    deleted_yn   BOOLEAN          NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS project_members (
    user_id    UUID        NOT NULL REFERENCES users(user_id)    ON DELETE CASCADE,
    project_id UUID        NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    role       TEXT        NOT NULL,
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_yn BOOLEAN     NOT NULL DEFAULT false,
    PRIMARY KEY (user_id, project_id)
);

CREATE TABLE IF NOT EXISTS workspaces (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID        NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(user_id)       ON DELETE CASCADE,
    status     TEXT        NOT NULL,
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_yn BOOLEAN     NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id         BIGSERIAL   PRIMARY KEY,
    user_id    UUID,
    action     TEXT        NOT NULL,
    resource   TEXT        NOT NULL,
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_yn BOOLEAN     NOT NULL DEFAULT false
);
`

const sqlIndexes = `
CREATE INDEX IF NOT EXISTS idx_users_deleted      ON users(deleted_yn);
CREATE INDEX IF NOT EXISTS idx_projects_deleted   ON projects(deleted_yn);
CREATE INDEX IF NOT EXISTS idx_workspaces_deleted ON workspaces(deleted_yn);
`
