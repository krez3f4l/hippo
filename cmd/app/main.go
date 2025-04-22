package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	grpcclient "hippo/internal/transport/grpc"

	"hippo/internal/platform/config"
	"hippo/internal/platform/database"
	"hippo/internal/platform/logger"
	"hippo/internal/repository/psql"
	"hippo/internal/service"
	"hippo/internal/transport/rest"
	"hippo/pkg/hash"
)

const (
	hashCost           = 10
	serverShutdownTime = 10 * time.Second
)

var (
	configDir  = os.Getenv("HIPPO_CONFIG_DIR")
	configName = os.Getenv("HIPPO_CONFIG_NAME")
	hmacSecret = os.Getenv("HIPPO_HMAC_SECRET")
)

func main() {
	cfg, err := config.NewConfig(configDir, configName)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log := logger.SetupLogrusLogger(cfg.Env)
	log.Info("Starting Medical CRUD API service", logger.String("env", cfg.Env))
	log.Debug("Enabled debug messages")

	db, err := database.NewPostgresConnection(cfg.DBConn)
	if err != nil {
		log.Fatal("failed to connect to database", logger.Err(err))
	}
	defer func() {
		if err = db.Close(); err != nil {
			log.Error("failed to close database connection", logger.Err(err))
		}
	}()

	log.Info("Database connection established",
		logger.String("host", cfg.DBConn.Host),
		logger.Int("port", cfg.DBConn.Port),
	)

	hasher := hash.NewBcryptHasher(hashCost)

	auditService, err := grpcclient.NewClient(cfg.GrpcAudit)
	if err != nil {
		log.Fatal("failed to init audit service", logger.Err(err))
	}
	defer func() {
		if err = auditService.Close(); err != nil {
			log.Error("failed to close audit service", logger.Err(err))
		}
	}()

	log.Info("gRPC audit server connection established",
		logger.String("host", cfg.GrpcAudit.Host),
		logger.Int("port", cfg.GrpcAudit.Port),
	)

	medicineService := service.NewMedicines(
		psql.NewMedicines(db),
		auditService,
		log,
	)

	usersService := service.NewUsers(
		psql.NewUsers(db),
		psql.NewToken(db),
		hasher,
		auditService,
		[]byte(hmacSecret),
		cfg.App.RefreshTokenLife,
		cfg.App.AccessTokenLife,
		log,
	)

	handler := rest.NewHandler(medicineService, usersService, log, cfg.App.HandlerTimeout)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HttpServer.Port),
		Handler:      handler.InitRouter(),
		ReadTimeout:  cfg.HttpServer.ReadTimeout,
		WriteTimeout: cfg.HttpServer.WriteTimeout,
		IdleTimeout:  cfg.HttpServer.Idle,
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("failed to connect to database", logger.Err(err))
		}
	}()

	log.Info("HTTP server started",
		logger.Int("port", cfg.HttpServer.Port),
	)

	<-shutdownChan
	log.Info("shutting down HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTime)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		log.Error("HTTP server shutdown error", logger.Err(err))
	}

	log.Info("HTTP server stopped")
}
