package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/howallet/howallet/internal/config"
	"github.com/howallet/howallet/internal/handler"
	"github.com/howallet/howallet/internal/repository/postgres"
	"github.com/howallet/howallet/internal/router"
	"github.com/howallet/howallet/internal/service"
)

func main() {
	// Load .env (ignore error — env vars might be set directly)
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Database connection pool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DB.DSN())
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("failed to ping database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("connected to database")

	// Repository layer
	repos := postgres.New(pool)

	// Services (repository-based)
	emailSvc := service.NewEmailService(&cfg.SMTP)
	authSvc := service.NewAuthService(repos, &cfg.JWT)
	hhSvc := service.NewHouseholdService(repos, emailSvc, cfg.Frontend.URL)
	accSvc := service.NewAccountService(repos.Accounts)
	txnSvc := service.NewTransactionService(repos)
	exportSvc := service.NewExportService(repos.Transactions)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	hhH := handler.NewHouseholdHandler(hhSvc)
	accH := handler.NewAccountHandler(accSvc)
	txnH := handler.NewTransactionHandler(txnSvc)
	expH := handler.NewExportHandler(exportSvc)

	// Router (membership check enforced in HouseholdCtx middleware)
	mux := router.New(cfg, logger, authH, hhH, accH, txnH, expH, hhSvc.CheckMembership)

	// HTTP Server
	srv := &http.Server{
		Addr:         cfg.API.Addr(),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		logger.Info("shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown error", slog.String("error", err.Error()))
		}
	}()

	logger.Info(fmt.Sprintf("hoWallet API listening on %s", cfg.API.Addr()))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("server stopped")
}
