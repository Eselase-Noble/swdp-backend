package main

// @title           SWDP Backend API
// @version         1.0
// @description     Web-Based Software Development Platform Backend
// @termsOfService  https://example.com/terms/

// @contact.name   Noble Eselase Vulley
// @contact.email  eselasenobleson@gmail.com

// @license.name  Proprietary
// @license.url   https://example.com/license

// @host      localhost:8282
// @BasePath  /api
// @schemes   http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"context"
	"log"
	"net/http"
	"os"
	_ "web-based-dev-platform-backend/docs"
	"web-based-dev-platform-backend/internal/auth"
	"web-based-dev-platform-backend/internal/config"
	"web-based-dev-platform-backend/internal/database"
	middleware2 "web-based-dev-platform-backend/internal/middleware"
	"web-based-dev-platform-backend/internal/projects"
	"web-based-dev-platform-backend/internal/users"
	"web-based-dev-platform-backend/internal/websocket"
	"web-based-dev-platform-backend/internal/workspaces"

	"github.com/docker/docker/client"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

var (
	serviceName = getServiceName()

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "swdp",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"service", "method", "path", "status"},
	)
)

func getServiceName() string {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	if s := os.Getenv("SERVICE_NAME"); s != "" {
		return s
	}
	return "unknown-service"
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

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

	// Prometheus
	prometheus.MustRegister(httpRequestsTotal)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Prometheus middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			httpRequestsTotal.WithLabelValues(
				serviceName,
				r.Method,
				r.URL.Path,
				http.StatusText(ww.Status()),
			).Inc()
		})
	})

	// Routes
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Route("/api", func(api chi.Router) {
		auth.RegisterAuth(api, db, cfg)

		// Secured routes
		api.Group(func(protected chi.Router) {
			protected.Use(middleware2.JWTAuth(cfg.JWTSecret))

			projects.RegisterProjects(protected, db)
			workspaces.RegisterWorkspaces(protected, db, cfg)
			users.RegisterUsers(protected, db, cfg)
			websocket.RegisterWebSockets(protected, dockerClient)
		})
	})

	log.Println("🚀 SWDP backend server running on :8282")
	err = http.ListenAndServe(":8282", r)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
