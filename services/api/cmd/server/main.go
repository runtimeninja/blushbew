package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/runtimeninja/blushbew/services/api/internal/auth"
	"github.com/runtimeninja/blushbew/services/api/internal/blog"
	"github.com/runtimeninja/blushbew/services/api/internal/config"
	"github.com/runtimeninja/blushbew/services/api/internal/db"
	"github.com/runtimeninja/blushbew/services/api/internal/httpserver"
	"github.com/runtimeninja/blushbew/services/api/internal/observability"

	// ✅ NEW IMPORTS
	"github.com/runtimeninja/blushbew/services/api/internal/ai"
	"github.com/runtimeninja/blushbew/services/api/internal/diagnosis"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	logger := observability.NewLogger(cfg.Env)

	// DB connect
	database, err := db.Connect(ctx, cfg.DBDSN)
	if err != nil {
		logger.Error("db connect failed", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Run migrations
	if err := db.Migrate(ctx, db.PoolAdaptor{Pool: database.Pool}); err != nil {
		logger.Error("db migrate failed", "error", err)
		os.Exit(1)
	}
	logger.Info("db migrations ok")

	// --- Domain wiring (Repo/Service) ---
	blogRepo := blog.NewRepo(database.Pool)
	authSvc := auth.New(database.Pool)

	// Seed admin user from env (only if not exists)
	if err := authSvc.EnsureAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		logger.Error("ensure admin failed", "error", err)
		os.Exit(1)
	}

	// ✅ NEW: AI + Diagnosis wiring
	aiClient := ai.NewOpenAI(cfg.OpenAIAPIKey, cfg.OpenAIModel, cfg.OpenAITimeout)
	diagSvc := diagnosis.NewService(aiClient)

	// --- Router wiring ---
	router := httpserver.NewRouter(httpserver.Deps{
		Logger:         logger,
		Pool:           database.Pool,
		AllowedOrigins: cfg.AllowedOrigins,
		Auth:           authSvc,
		Blog:           blogRepo,
		Diagnosis:      diagSvc,
	})

	// --- HTTP server ---
	srv := httpserver.New(httpserver.ServerDeps{
		Logger:       logger,
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	})

	// --- Graceful shutdown ---
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			logger.Error("http server error", "error", err)
		}
	}()

	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	_ = srv.Shutdown(shutdownCtx)
	time.Sleep(200 * time.Millisecond)
	logger.Info("shutdown complete")
}
