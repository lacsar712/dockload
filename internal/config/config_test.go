package config_test

import (
	"testing"

	"github.com/lacsar712/dockload/internal/config"
)

func TestDefaultValidate(t *testing.T) {
	cfg := config.Default()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateWindowSize(t *testing.T) {
	cfg := config.Default()
	cfg.WindowSize = 1
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestStableConfig(t *testing.T) {
	cfg := config.Default()
	n, eps := cfg.StableConfig()
	if n != 10 || eps != 5.0 {
		t.Fatalf("unexpected stable config %d %f", n, eps)
	}
}
