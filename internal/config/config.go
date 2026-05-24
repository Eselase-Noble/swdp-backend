package config

import (
	"log"
	"os"

	"github.com/docker/docker/client"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Env          string
	DBUrl        string
	JWTSecret    string
	DockerHost   string
	Port         string
	ServiceName  string
	DockerClient *client.Client

	// SeedOnStart inserts dev seed users on every startup if true.
	// Defaults to true when ENV=development.
	SeedOnStart bool

	// RuntimeType selects the workspace execution backend.
	// "docker"      → Docker containers (default, works everywhere Docker is installed)
	// "firecracker" → Firecracker MicroVMs (requires KVM; see internal/runtime/firecracker.go)
	RuntimeType string

	// Firecracker-specific — only needed when RuntimeType = "firecracker"
	FirecrackerKernelPath   string // path to uncompressed vmlinux binary
	FirecrackerRootfsBase   string // path to the base ext4 rootfs template
	FirecrackerWorkspaceDir string // directory for per-workspace VM files
}

// Load reads configuration from environment variables.
func Load() *Config {
	env := get("ENV", "development")
	cfg := &Config{
		Env:         env,
		DBUrl:       must("DATABASE_URL"),
		JWTSecret:   must("JWT_SECRET"),
		DockerHost:  get("DOCKER_HOST", "unix:///var/run/docker.sock"),
		Port:        get("PORT", "8282"),
		ServiceName: get("SERVICE_NAME", "swdp-backend"),

		// Seed by default in development; set SEED_ON_START=false to disable.
		SeedOnStart: getBool("SEED_ON_START", env == "development"),

		RuntimeType: get("RUNTIME_TYPE", "docker"),

		FirecrackerKernelPath:   get("FIRECRACKER_KERNEL_PATH", "/opt/firecracker/vmlinux"),
		FirecrackerRootfsBase:   get("FIRECRACKER_ROOTFS_BASE", "/opt/firecracker/ubuntu-22.04.ext4"),
		FirecrackerWorkspaceDir: get("FIRECRACKER_WORKSPACE_DIR", "/var/swdp/workspaces"),
	}
	return cfg
}

func must(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s not set", key)
	}
	return v
}

func get(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v == "true" || v == "1" || v == "yes"
}
