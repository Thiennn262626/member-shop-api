package main

import (
	"context"
	"member-shop-api/internal/app"
	"member-shop-api/internal/config"
	"member-shop-api/internal/utils"
	"member-shop-api/pkg/logger"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

func main() {
	rootDir := utils.MustGetWorkingDir()
	logFile := filepath.Join(rootDir, "./log/app.log")

	logger.InitLogger(logger.LoggerConfig{
		Level:      zerolog.InfoLevel,
		Filename:   logFile,
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     5,
		Compress:   true,
		IsDev:      utils.GetEnv("APP_ENV", "production"),
	})
	// Chỉ load file .env nếu APP_ENV chưa được set
	if utils.GetEnv("APP_ENV", "development") == "development" {
		if err := godotenv.Load(filepath.Join(rootDir, ".env")); err != nil {
			logger.Log.Warn().Msg("⚠️ No .env file found")
		} else {
			logger.Log.Info().Msg("✅ Loaded successfully .env in api proccess")
		}
	}

	// Initialize configuration
	cfg := config.NewConfig()

	// Initialize application
	application, err := app.NewApplication(cfg)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to initialize application")
	}

	// context quản lý lifecycle
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	// HTTP server
	g.Go(func() error {
		if err := application.Run(ctx); err != nil {
			logger.Log.Fatal().Err(err).Msg("Application run failed")
		}
		return nil
	})

	// wait
	if err := g.Wait(); err != nil {
		logger.Log.Warn().Err(err).Msg("⚠️ failed to stop application")
	}
}
