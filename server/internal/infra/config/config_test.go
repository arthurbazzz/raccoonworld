package config_test

import (
	"reflect"
	"testing"

	"raccoons-world/server/internal/infra/config"
)

func TestDefaults(t *testing.T) {
	t.Setenv("RW_ADDR", "")
	t.Setenv("RW_ALLOWED_ORIGINS", "")
	t.Setenv("RW_DEBUG", "")

	got := config.FromEnv()
	want := config.Config{Addr: ":8080", AllowedOrigins: []string{"localhost:3000"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FromEnv() = %+v, want %+v", got, want)
	}
}

func TestFromEnv(t *testing.T) {
	t.Setenv("RW_ADDR", "127.0.0.1:9000")
	t.Setenv("RW_ALLOWED_ORIGINS", " localhost:3000, raccoons.world ,, ")
	t.Setenv("RW_DEBUG", "1")

	got := config.FromEnv()
	want := config.Config{
		Addr:           "127.0.0.1:9000",
		AllowedOrigins: []string{"localhost:3000", "raccoons.world"},
		Debug:          true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FromEnv() = %+v, want %+v", got, want)
	}
}
