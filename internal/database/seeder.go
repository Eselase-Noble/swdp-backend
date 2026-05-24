package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type seedUser struct {
	Username string
	Name     string
	Email    string
	Password string
	Role     string
}

// devUsers is the fixed set of well-known accounts created on first run.
// Passwords are hashed by Go's bcrypt at seed time — pgcrypto is not required.
var devUsers = []seedUser{
	{"superadmin", "Super Admin", "admin@swdp.io", "Admin@1234", "admin"},
	{"nobleson", "Noble Eselase", "nobleson@swdp.io", "Noble@1234", "admin"},
	{"alice", "Alice Johnson", "alice@swdp.io", "Alice@1234", "developer"},
	{"bob", "Bob Smith", "bob@swdp.io", "Bob@1234", "developer"},
	{"carol", "Carol Williams", "carol@swdp.io", "Carol@1234", "developer"},
	{"david", "David Brown", "david@swdp.io", "David@1234", "developer"},
}

// Seed inserts the dev users and a demo project if they do not already exist.
// Controlled by the SEED_ON_START env var — see config.Config.SeedOnStart.
// All statements use ON CONFLICT DO NOTHING, so re-running is safe.
func Seed(ctx context.Context, pool *pgxpool.Pool) {
	log.Println("seeder: starting dev seed …")

	insertedCount := 0
	for _, u := range devUsers {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("seeder: bcrypt error for %s: %v", u.Email, err)
			continue
		}

		tag, err := pool.Exec(ctx, `
			INSERT INTO users (username, name, email, password_hash, role)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (email) DO NOTHING
		`, u.Username, u.Name, u.Email, string(hash), u.Role)
		if err != nil {
			log.Printf("seeder: insert user %s: %v", u.Email, err)
			continue
		}
		if tag.RowsAffected() > 0 {
			insertedCount++
			log.Printf("seeder: created user %-12s (%s)", u.Username, u.Role)
		}
	}

	// Create a demo project owned by the admin account.
	_, err := pool.Exec(ctx, `
		INSERT INTO projects (project_name, owner_id, version, environment)
		SELECT 'SWDP Demo', user_id, 1, 'dev'::environment_enum
		FROM   users
		WHERE  email = 'admin@swdp.io'
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		log.Printf("seeder: insert demo project: %v", err)
	}

	// Add alice and bob as members of the demo project.
	_, err = pool.Exec(ctx, `
		INSERT INTO project_members (user_id, project_id, role)
		SELECT u.user_id, p.project_id, 'contributor'
		FROM   users    u
		JOIN   projects p ON p.project_name = 'SWDP Demo'
		WHERE  u.email IN ('alice@swdp.io', 'bob@swdp.io')
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		log.Printf("seeder: insert project members: %v", err)
	}

	if insertedCount > 0 {
		log.Printf("seeder: inserted %d new user(s)", insertedCount)
	} else {
		log.Println("seeder: all seed users already exist — nothing to do")
	}
	log.Println("seeder: done")
	log.Println("seeder: ─────────────────────────────────────────")
	log.Println("seeder:  admin@swdp.io      →  Admin@1234  (admin)")
	log.Println("seeder:  nobleson@swdp.io   →  Noble@1234  (admin)")
	log.Println("seeder:  alice@swdp.io      →  Alice@1234  (developer)")
	log.Println("seeder:  bob@swdp.io        →  Bob@1234    (developer)")
	log.Println("seeder:  carol@swdp.io      →  Carol@1234  (developer)")
	log.Println("seeder:  david@swdp.io      →  David@1234  (developer)")
	log.Println("seeder: ─────────────────────────────────────────")
}
