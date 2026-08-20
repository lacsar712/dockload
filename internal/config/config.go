// Package config loads runtime settings for dockload from environment variables.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all tunable parameters for the weighing pipeline.
type Config struct {
	Addr           string
	WindowSize     int
	StableEps      float64
	LoadChange     float64
	MaxRecent      int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	ShutdownGrace  time.Duration
	DefaultMaxLoad float64
}

// Default returns production-safe defaults aligned with SPEC.
func Default() Config {
	return Config{
		Addr:           ":8080",
		WindowSize:     10,
		StableEps:      5.0,
		LoadChange:     50.0,
		MaxRecent:      100,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		ShutdownGrace:  15 * time.Second,
		DefaultMaxLoad: 45000,
	}
}

// Load merges environment overrides onto defaults.
func Load() Config {
	cfg := Default()
	if v := os.Getenv("DOCKLOAD_ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := os.Getenv("DOCKLOAD_WINDOW_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.WindowSize = n
		}
	}
	if v := os.Getenv("DOCKLOAD_STABLE_EPS"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 {
			cfg.StableEps = f
		}
	}
	if v := os.Getenv("DOCKLOAD_LOAD_CHANGE"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			cfg.LoadChange = f
		}
	}
	if v := os.Getenv("DOCKLOAD_MAX_RECENT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MaxRecent = n
		}
	}
	return cfg
}

// StableConfig extracts stable-window parameters.
func (c Config) StableConfig() (windowSize int, eps float64) {
	return c.WindowSize, c.StableEps
}

// WeighConfig extracts FSM load-change threshold.
func (c Config) WeighConfig() float64 {
	return c.LoadChange
}
