package main

import (
	"context"
	"log"
	"net/http"
	"web-based-dev-platform-backend/internal/auth"
	"web-based-dev-platform-backend/internal/config"
	"web-based-dev-platform-backend/internal/database"
	"web-based-dev-platform-backend/internal/projects"
	"web-based-dev-platform-backend/internal/users"
	"web-based-dev-platform-backend/internal/websocket"
	"web-based-dev-platform-backend/internal/workspaces"

	"github.com/docker/docker/client"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg.DBUrl)
	//database.RunMigrations(db)

	// Initialize Docker client
	dockerClient, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Fatal("Failed to create Docker client:", err)
	}
	defer func(dockerClient *client.Client) {
		err := dockerClient.Close()
		if err != nil {
			return
		}
	}(dockerClient)

	// Test Docker connection
	_, err = dockerClient.Ping(context.Background())
	if err != nil {
		log.Fatal("Cannot connect to Docker daemon. Make sure Docker is running:", err)
	}
	log.Println("✅ Successfully connected to Docker daemon")

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(api chi.Router) {
		auth.RegisterAuth(api, db, cfg)
		projects.RegisterProjects(api, db)
		workspaces.RegisterWorkspaces(api, db, cfg)
		websocket.RegisterWebSockets(api, dockerClient)
		users.RegisterUsers(api, db, cfg)
	})

	log.Println("🚀 SWDP backend server running on :8282")
	err = http.ListenAndServe(":8282", r)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
