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

// loginRequest represents the JSON body for login
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterAuth registers login and logout routes
func RegisterAuth(r chi.Router, db *pgxpool.Pool, cfg *config.Config) {
	r.Post("/auth/login", login(db, cfg))
	r.Post("/auth/logout", logout())
}

// login handles user authentication
func login(db *pgxpool.Pool, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}(r.Body)

		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		var userId, role, passwordHash string

		err := db.QueryRow(
			r.Context(),
			"SELECT userId, role, password FROM users WHERE email = $1",
			req.Email,
		).Scan(&userId, &role, &passwordHash)

		// Always return same error message for wrong email or password
		if err != nil || !CheckPassword(req.Password, passwordHash) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		token, err := Generate(userId, role, cfg.JWTSecret)
		if err != nil {
			http.Error(w, "Token generation error", http.StatusInternalServerError)
			return
		}

		// Set the JWT token as a secure, HttpOnly cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			HttpOnly: true,
			Secure:   true, // true in production
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		})

		w.WriteHeader(http.StatusOK)
	}
}

// logout clears the authentication cookie
func logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true, // match login cookie
			SameSite: http.SameSiteStrictMode,
		})
		w.WriteHeader(http.StatusOK)
	}
}

// CheckPassword verifies a plaintext password against a bcrypt hash
func CheckPassword(password, passwordHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	return err == nil
}

// HashPassword generates a bcrypt hash from a plaintext password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
