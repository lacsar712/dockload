package config

import "fmt"

// Validate checks that configuration values are usable at startup.
func (c Config) Validate() error {
	if c.WindowSize < 2 {
		return fmt.Errorf("window size must be at least 2, got %d", c.WindowSize)
	}
	if c.StableEps < 0 {
		return fmt.Errorf("stable epsilon must be non-negative, got %f", c.StableEps)
	}
	if c.LoadChange <= 0 {
		return fmt.Errorf("load change threshold must be positive, got %f", c.LoadChange)
	}
	if c.MaxRecent < 1 {
		return fmt.Errorf("max recent events must be at least 1, got %d", c.MaxRecent)
	}
	if c.Addr == "" {
		return fmt.Errorf("listen address must not be empty")
	}
	return nil
}

// String returns a concise summary for logging.
func (c Config) String() string {
	return fmt.Sprintf("addr=%s window=%d eps=%.2f loadChange=%.2f maxRecent=%d",
		c.Addr, c.WindowSize, c.StableEps, c.LoadChange, c.MaxRecent)
}
