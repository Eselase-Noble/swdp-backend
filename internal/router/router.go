package router

import (
	"net/http"
	"web-based-dev-platform-backend/internal/auth"
	"web-based-dev-platform-backend/internal/config"
	"web-based-dev-platform-backend/internal/middleware"
	"web-based-dev-platform-backend/internal/projects"
	"web-based-dev-platform-backend/internal/users"
	"web-based-dev-platform-backend/internal/websocket"
	"web-based-dev-platform-backend/internal/workspaces"

	"github.com/docker/docker/client"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"
)

type Deps struct {
	Config       *config.Config
	SQLDB        *pgxpool.Pool
	GormDB       *gorm.DB
	DockerClient *client.Client
}

//var (
//	serviceName = getServiceName()
//
//	httpRequestsTotal = prometheus.NewCounterVec(
//		prometheus.CounterOpts{
//			Namespace: "swdp",
//			Subsystem: "http",
//			Name:      "requests_total",
//			Help:      "Total number of HTTP requests",
//		},
//		[]string{"service", "method", "path", "status"},
//	)
//)

//func getServiceName() string {
//	err := godotenv.Load()
//	if err != nil {
//		log.Println("No .env file found, using system environment variables")
//	}
//	if s := os.Getenv("SERVICE_NAME"); s != "" {
//		return s
//	}
//	return "unknown-service"
//}

func New(d Deps) http.Handler {

	// Prometheus
	//prometheus.MustRegister(httpRequestsTotal)

	r := chi.NewRouter()

	// ======================
	// Global middleware
	// ======================
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	// Prometheus middleware
	//r.Use(func(next http.Handler) http.Handler {
	//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
	//		next.ServeHTTP(ww, r)
	//
	//		httpRequestsTotal.WithLabelValues(
	//			serviceName,
	//			r.Method,
	//			r.URL.Path,
	//			http.StatusText(ww.Status()),
	//		).Inc()
	//	})
	//})

	// ======================
	// Public routes
	// ======================
	r.Get("/health", health)
	r.Get("/metrics", promhttp.Handler().ServeHTTP)
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// ======================
	// API routes
	// ======================
	r.Route("/swdp/v1/api", func(api chi.Router) {

		// Auth (public)
		auth.RegisterAuth(api, d.SQLDB, d.Config)

		// Protected
		api.Group(func(protected chi.Router) {
			protected.Use(middleware.JWTAuth(d.Config.JWTSecret))

			// Users
			userRepo := &users.Repository{DB: d.GormDB}
			userService := &users.Service{Repo: userRepo}
			userHandler := users.NewHandler(userService)
			users.UserRoutes(protected, userHandler)

			// Other modules
			projects.RegisterProjects(protected, d.SQLDB)
			workspaces.RegisterWorkspaces(protected, d.SQLDB, d.Config)
			websocket.RegisterWebSockets(protected, d.DockerClient)
		})
	})

	return r
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
