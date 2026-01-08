package config

import (
	"log"
	"os"

	"github.com/docker/docker/client"
)

// Interface for the Config Struct
type Config struct {
	Env          string
	DBUrl        string
	JWTSecret    string
	DockerHost   string
	DockerClient *client.Client
}

// Load the configuration from the environment variable
func Load() *Config {
	cfg := &Config{
		Env:        get("ENV", "development"),
		DBUrl:      must("DATABASE_URL"),
		JWTSecret:  must("JWT_SECRET"),
		DockerHost: get("DOCKER_HOST", "unix:///var/run/docker.sock"),
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
