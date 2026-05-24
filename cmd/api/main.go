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
	"web-based-dev-platform-backend/internal/runtime"

	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	sqlDB := database.Connect(cfg.DBUrl)
	gormDB := database.ConnectDB(cfg)

	// Select and initialise the workspace execution runtime.
	rt := buildRuntime(cfg)

	r := router.New(router.Deps{
		Config:  cfg,
		SQLDB:   sqlDB,
		GormDB:  gormDB,
		Runtime: rt,
	})

	log.Printf("SWDP backend running on :%s  (runtime: %s)\n", cfg.Port, cfg.RuntimeType)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}

func buildRuntime(cfg *config.Config) runtime.Runtime {
	switch cfg.RuntimeType {
	case "firecracker":
		log.Println("runtime: Firecracker MicroVM")
		fc, err := runtime.NewFirecrackerRuntime(runtime.FirecrackerConfig{
			KernelPath:   cfg.FirecrackerKernelPath,
			RootfsBase:   cfg.FirecrackerRootfsBase,
			WorkspaceDir: cfg.FirecrackerWorkspaceDir,
		})
		if err != nil {
			log.Fatal("firecracker runtime init:", err)
		}
		return fc

	default:
		log.Println("runtime: Docker")
		dockerClient, err := client.NewClientWithOpts(
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			log.Fatal("docker client:", err)
		}
		if _, err := dockerClient.Ping(context.Background()); err != nil {
			log.Fatal("docker not reachable:", err)
		}
		return runtime.NewDockerRuntime(dockerClient)
	}
}
