package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/somya/git-banner-backend/internal/config"
	"github.com/somya/git-banner-backend/internal/server"
)

const asciiBanner = `
` + "\x1b[32m\x1b[1m" + ` ██████╗ ██╗████████╗██████╗  █████╗ ███╗  ██╗███╗  ██╗███████╗██████╗
██╔════╝ ██║╚══██╔══╝██╔══██╗██╔══██╗████╗ ██║████╗ ██║██╔════╝██╔══██╗
██║  ███╗██║   ██║   ██████╔╝███████║██╔██╗██║██╔██╗██║█████╗  ██████╔╝
██║   ██║██║   ██║   ██╔══██╗██╔══██║██║╚████║██║╚████║██╔══╝  ██╔══██╗
╚██████╔╝██║   ██║   ██████╔╝██║  ██║██║ ╚███║██║ ╚███║███████╗██║  ██║
 ╚═════╝ ╚═╝   ╚═╝   ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚══╝╚═╝  ╚══╝╚══════╝╚═╝  ╚═╝` + "\x1b[0m" + `
` + "\x1b[2m" + `  GitHub Profile Banner Generator — Backend Server` + "\x1b[0m" + `
`

func main() {
	fmt.Print(asciiBanner)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	srv := server.New(cfg)

	go func() {
		slog.Info("git-banner-backend starting", "port", cfg.Port, "env", cfg.Env)
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
