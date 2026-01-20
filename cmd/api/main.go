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
// @BasePath  /swdp/v1/api
// @schemes   http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
import (
	"context"
	"log"
	"net/http"

	_ "web-based-dev-platform-backend/docs"

	"web-based-dev-platform-backend/internal/config"
	"web-based-dev-platform-backend/internal/database"
	"web-based-dev-platform-backend/internal/router"

	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	// DBs
	sqlDB := database.Connect(cfg.DBUrl)
	gormDB := database.ConnectDB(cfg)

	// Docker
	dockerClient, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer dockerClient.Close()

	if _, err := dockerClient.Ping(context.Background()); err != nil {
		log.Fatal("Docker not reachable:", err)
	}

	// Router
	r := router.New(router.Deps{
		Config:       cfg,
		SQLDB:        sqlDB,
		GormDB:       gormDB,
		DockerClient: dockerClient,
	})

	log.Println("🚀 SWDP backend running on :8282")
	log.Fatal(http.ListenAndServe(":8282", r))
}
