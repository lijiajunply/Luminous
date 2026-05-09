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
	"luminous/internal/school/xauat"

	"github.com/gin-gonic/gin"
)

func main() {
	slog.Info("Starting Luminous server")

	if err := config.LoadConfig(); err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	gin.SetMode(config.Cfg.Server.Mode)

	xauat.Init(config.Cfg.Schools.XAUAT)

	repo, err := repository.NewJSONSchoolRepository(config.Cfg.Data.SchoolsFile)
	if err != nil {
		slog.Error("Failed to initialize repository", "error", err)
		os.Exit(1)
	}

	schoolHandler := handler.NewSchoolHandler(repo)
	adminHandler := handler.NewAdminHandler(repo)
	xauatHandler := handler.NewXAUATHandler()

	r := router.SetupRouter(schoolHandler, adminHandler, xauatHandler)

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	xauat.Shutdown()
	slog.Info("Server exited")
}
