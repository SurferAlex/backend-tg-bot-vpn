package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-vpn/internal/api"
	"api-vpn/internal/config"
	"api-vpn/internal/repository"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := config.LoadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := repository.Open(ctx, cfg.DatabaseDNS)
	if err != nil {
		logger.Error("open psql", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	serversRepo := repository.NewVPNServersRepo(pool)
	if err := repository.SyncDefaultServerFromEnv(ctx, serversRepo, cfg.XUI); err != nil {
		logger.Error("sync default vpn server from env", "error", err)
		os.Exit(1)
	}
	if err := repository.EnsureOptionalServersFromEnv(ctx, serversRepo, cfg.OptionalServers); err != nil {
		logger.Error("bootstrap optional vpn servers from env", "error", err)
		os.Exit(1)
	}

	router := api.SetupServer(pool, cfg.InternalToken)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownContext); err != nil {
			logger.Error("shutdown http server", "error", err)
		}
	}()

	logger.Info("http server started", "addr", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("listen and serve", "error", err)
		os.Exit(1)
	}

}
