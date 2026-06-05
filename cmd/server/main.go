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

	"luminous/internal/config"
	"luminous/internal/handler"
	"luminous/internal/repository"
	"luminous/internal/router"

	"github.com/gin-gonic/gin"
)

func main() {
	slog.Info("Starting Luminous server")

	if err := config.LoadConfig(); err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	gin.SetMode(config.Cfg.Server.Mode)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pgRepo, err := repository.NewPGSchoolRepository(ctx, config.Cfg.Database)
	if err != nil {
		slog.Error("Failed to initialize PostgreSQL repository", "error", err)
		os.Exit(1)
	}
	defer pgRepo.Close()

	schoolHandler := handler.NewSchoolHandler(pgRepo)
	adminHandler := handler.NewAdminHandler(pgRepo)
	appHandler := handler.NewAppHandler()

	r := router.SetupRouter(schoolHandler, adminHandler, appHandler)

	addr := fmt.Sprintf(":%d", config.Cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited")
}
