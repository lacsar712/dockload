// Command dockload runs the container terminal weigh-stream HTTP service.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lacsar712/dockload/internal/app"
	"github.com/lacsar712/dockload/internal/config"
)

func main() {
	addr := flag.String("addr", "", "listen address (overrides DOCKLOAD_ADDR)")
	flag.Parse()

	cfg := config.Load()
	if *addr != "" {
		cfg.Addr = *addr
	}

	logger := app.DefaultLogger()
	application, err := app.New(cfg, logger)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := application.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("run: %v", err)
	}
	logger.Println("dockload stopped")
}
