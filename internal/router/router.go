package router

import (
	"net/http"
	"web-based-dev-platform-backend/internal/auth"
	"web-based-dev-platform-backend/internal/config"
	"web-based-dev-platform-backend/internal/middleware"
	"web-based-dev-platform-backend/internal/projectmembers"
	"web-based-dev-platform-backend/internal/projects"
	"web-based-dev-platform-backend/internal/runtime"
	"web-based-dev-platform-backend/internal/users"
	"web-based-dev-platform-backend/internal/websocket"
	"web-based-dev-platform-backend/internal/workspaces"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"

	"github.com/go-chi/chi/v5"
)

type Deps struct {
	Config  *config.Config
	SQLDB   *pgxpool.Pool
	GormDB  *gorm.DB
	Runtime runtime.Runtime // Docker or Firecracker — selected at startup
}

func New(d Deps) http.Handler {
	RegisterMetrics()

	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(PrometheusMiddleware)

	// Public
	r.Get("/health", health)
	r.Get("/metrics", promhttp.Handler().ServeHTTP)
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/swdp/v1/api", func(api chi.Router) {
		auth.RegisterAuth(api, d.SQLDB, d.Config)

		api.Group(func(protected chi.Router) {
			protected.Use(middleware.JWTAuth(d.Config.JWTSecret))

			// Users
			userRepo := &users.Repository{DB: d.GormDB}
			userService := &users.Service{Repo: userRepo}
			userHandler := users.NewHandler(userService)
			users.UserRoutes(protected, userHandler)

			// Projects
			projectRepo := &projects.Repository{DB: d.GormDB}
			projectService := &projects.Service{Repo: projectRepo}
			projectHandler := projects.NewHandler(projectService)
			projects.RegisterRoutes(protected, projectHandler)

			// Project Members
			projectMemberRepo := &projectmembers.Repository{DB: d.GormDB}
			projectMemberService := &projectmembers.Service{Repo: projectMemberRepo}
			projectMemberHandler := projectmembers.NewHandler(projectMemberService)
			projectmembers.RegisterRoutes(protected, projectMemberHandler)

			// Workspaces — backed by the selected runtime
			workSpaceRepo := &workspaces.Repository{DB: d.GormDB}
			workSpaceService := workspaces.NewService(workSpaceRepo, d.Runtime)
			workSpaceHandler := workspaces.NewHandler(workSpaceService)
			workspaces.Routes(protected, workSpaceHandler)

		})

		// WebSocket routes use ?token= auth — browsers can't send headers during WS upgrade
		api.Group(func(wsGroup chi.Router) {
			wsGroup.Use(middleware.JWTAuthWS(d.Config.JWTSecret))
			websocket.RegisterWebSockets(wsGroup, d.Runtime)
		})
	})

	return r
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
