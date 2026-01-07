package main

import (
	"log"
	"net/http"
	"web-based-dev-platform-backend/internal/auth"
	"web-based-dev-platform-backend/internal/config"
	"web-based-dev-platform-backend/internal/database"
	"web-based-dev-platform-backend/internal/projects"
	"web-based-dev-platform-backend/internal/websocket"
	"web-based-dev-platform-backend/internal/workspaces"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {
	cfg := config.Load()
	db := database.Connect(cfg.DBUrl)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(api chi.Router) {
		auth.RegisterAuth(api, db, cfg)
		projects.RegisterProjects(api, db)
		workspaces.RegisterWorkspaces(api, db, cfg)
		websocket.RegisterWebSockets(api, db, cfg)
	})

	log.Println("SWDP baceknd server running on :8282 ")
	err := http.ListenAndServe(":8282", r)
	if err != nil {
		return
	}

}
