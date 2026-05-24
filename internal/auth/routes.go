package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"web-based-dev-platform-backend/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// RegisterAuth registers public auth routes (no JWT required).
func RegisterAuth(r chi.Router, db *pgxpool.Pool, cfg *config.Config) {
	r.Post("/auth/login", login(db, cfg))
	r.Post("/auth/register", register(db))
	r.Post("/auth/logout", logout())
}

// login authenticates a user and sets a JWT cookie.
func login(db *pgxpool.Pool, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		var userID, role, passwordHash string
		err := db.QueryRow(
			r.Context(),
			// Column names match migrations/001_init.sql exactly.
			"SELECT user_id, role, password_hash FROM users WHERE email = $1 AND deleted_yn = false",
			req.Email,
		).Scan(&userID, &role, &passwordHash)

		if err != nil || !CheckPassword(req.Password, passwordHash) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		token, err := Generate(userID, role, cfg.JWTSecret)
		if err != nil {
			http.Error(w, "token generation error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}

// register creates a new developer account.
func register(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func(body io.ReadCloser) { body.Close() }(r.Body)

		var req registerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if req.Email == "" || req.Password == "" || req.Username == "" || req.Name == "" {
			http.Error(w, "username, name, email and password are required", http.StatusBadRequest)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		role := req.Role
		if role == "" {
			role = "developer"
		}

		var userID string
		err = db.QueryRow(
			r.Context(),
			`INSERT INTO users (username, name, email, password_hash, role)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING user_id`,
			req.Username, req.Name, req.Email, string(hash), role,
		).Scan(&userID)

		if err != nil {
			http.Error(w, "registration failed — email or username may already be taken", http.StatusConflict)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"user_id": userID})
	}
}

// logout clears the auth cookie.
func logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		w.WriteHeader(http.StatusOK)
	}
}

// CheckPassword verifies a plaintext password against a bcrypt hash.
func CheckPassword(password, passwordHash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}

// HashPassword generates a bcrypt hash from a plaintext password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
