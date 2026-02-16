package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	os.Unsetenv("SERVER_ADDR")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("REDIS_ADDR")

	cfg := Load()

	if cfg.ServerAddr != ":8080" {
		t.Errorf("got ServerAddr %q, want :8080", cfg.ServerAddr)
	}
	if cfg.RedisAddr != "localhost:6379" {
		t.Errorf("got RedisAddr %q, want localhost:6379", cfg.RedisAddr)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	os.Setenv("SERVER_ADDR", ":9090")
	os.Setenv("REDIS_ADDR", "redis:6380")
	defer os.Unsetenv("SERVER_ADDR")
	defer os.Unsetenv("REDIS_ADDR")

	cfg := Load()

	if cfg.ServerAddr != ":9090" {
		t.Errorf("got ServerAddr %q, want :9090", cfg.ServerAddr)
	}
	if cfg.RedisAddr != "redis:6380" {
		t.Errorf("got RedisAddr %q, want redis:6380", cfg.RedisAddr)
	}
}
