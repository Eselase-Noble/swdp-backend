package users

import (
	"encoding/json"
	"net/http"
	"web-based-dev-platform-backend/internal/auth"
	"web-based-dev-platform-backend/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRequest struct {
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"`
}

type userResponse struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

func RegisterUsers(r chi.Router, db *pgxpool.Pool, cfg *config.Config) {
	// User management
	r.Post("/users/create-user", createUser(db))
	r.Put("/users/{userId}", updateUser(db))
	r.Delete("/users/{userId}", deleteUser(db))
	r.Get("/users/all", getUsers(db))
	r.Get("/users/{userId}", getUser(db))
}

func createUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req userRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if req.Email == "" || req.Password == "" || req.Role == "" {
			http.Error(w, "Missing fields", http.StatusBadRequest)
			return
		}

		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			http.Error(w, "Password error", http.StatusInternalServerError)
			return
		}

		var userId string
		err = db.QueryRow(
			r.Context(),
			`INSERT INTO users (userId, email, password, role)
			 VALUES (gen_random_uuid(), $1, $2, $3)
			 RETURNING userId`,
			req.Email, hash, req.Role,
		).Scan(&userId)

		if err != nil {
			http.Error(w, "User creation failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(userResponse{
			UserID: userId,
			Email:  req.Email,
			Role:   req.Role,
		})
	}
}

func updateUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := chi.URLParam(r, "userId")

		var req userRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		//if req.Password != "" {
		//	hash, err := auth.HashPassword(req.Password)
		//	if err != nil {
		//		http.Error(w, "Password error", http.StatusInternalServerError)
		//		return
		//	}
		//
		//	_, err = db.Exec(
		//		r.Context(),
		//		`UPDATE users SET email=$1, password=$2, role=$3 WHERE userId=$4`,
		//		req.Email, hash, req.Role, userId,
		//	)
		//	if err != nil {
		//		http.Error(w, "Update failed", http.StatusInternalServerError)
		//		return
		//	}
		//} else {
		_, err := db.Exec(
			r.Context(),
			`UPDATE users SET email=$1, role=$2 WHERE userId=$3`,
			req.Email, req.Role, userId,
		)
		if err != nil {
			http.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func deleteUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := chi.URLParam(r, "userId")

		_, err := db.Exec(
			r.Context(),
			"DELETE FROM users WHERE userId = $1",
			userId,
		)
		if err != nil {
			http.Error(w, "Delete failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func getUsers(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(
			r.Context(),
			"SELECT user_id, email, role FROM users",
		)
		if err != nil {
			http.Error(w, "Query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []userResponse
		for rows.Next() {
			var u userResponse
			if err := rows.Scan(&u.UserID, &u.Email, &u.Role); err != nil {
				http.Error(w, "Scan failed", http.StatusInternalServerError)
				return
			}
			users = append(users, u)
		}

		json.NewEncoder(w).Encode(users)
	}
}

func getUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := chi.URLParam(r, "userId")

		var u userResponse
		err := db.QueryRow(
			r.Context(),
			"SELECT userId, email, role FROM users WHERE userId = $1",
			userId,
		).Scan(&u.UserID, &u.Email, &u.Role)

		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(u)
	}
}
