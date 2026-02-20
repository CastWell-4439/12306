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

	"github.com/gin-gonic/gin"

	commonconfig "ticketing/internal/common/config"
	commonkafka "ticketing/internal/common/kafka"
	"ticketing/internal/common/logging"
	commonmetrics "ticketing/internal/common/metrics"
	"ticketing/internal/common/middleware"
	commonmysql "ticketing/internal/common/mysql"
	commonredis "ticketing/internal/common/redis"
	"ticketing/internal/inventory/application"
	"ticketing/internal/inventory/infrastructure/event"
	"ticketing/internal/inventory/infrastructure/snapshot"
	"ticketing/internal/inventory/infrastructure/ttl"
	"ticketing/internal/inventory/infrastructure/wal"
	inventoryhttp "ticketing/internal/inventory/interfaces/http"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("inventory-service stopped with error: %v", err)
	}
}

func run() error {
	cfg := commonconfig.Load("inventory-service")
	logger := logging.New(cfg.ServiceName, cfg.Env, cfg.Version)

	mysqlDB, err := commonmysql.New(cfg.MySQLDSN)
	if err != nil {
		return fmt.Errorf("mysql init failed: %w", err)
	}
	defer mysqlDB.Close()

	redisClient, err := commonredis.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		return fmt.Errorf("redis init failed: %w", err)
	}
	defer redisClient.Close()

	kafkaProducer := commonkafka.NewProducer(cfg.KafkaBrokers)
	defer kafkaProducer.Close()
	if err := commonkafka.HealthCheck(context.Background(), cfg.KafkaBrokers); err != nil {
		return fmt.Errorf("kafka init failed: %w", err)
	}

	walRepo := wal.NewRepository(mysqlDB)
	snapshotRepo := snapshot.NewRepository(mysqlDB)
	publisher := event.NewPublisher(kafkaProducer, "inventory.events")
	holdStore := ttl.NewStore(redisClient, time.Duration(cfg.InventoryHoldTTLSecs)*time.Second)

	svc := application.NewService(
		logger,
		walRepo,
		snapshotRepo,
		publisher,
		holdStore,
		application.Config{
			ShardCount:           cfg.InventoryShardCount,
			WALBuffer:            cfg.InventoryWALBuffer,
			SnapshotInterval:     time.Duration(cfg.InventorySnapshotIntervalSecs) * time.Second,
			SnapshotOpsThreshold: cfg.InventorySnapshotOpsThreshold,
		},
	)
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := svc.Start(rootCtx); err != nil {
		return err
	}

	metrics := commonmetrics.New(cfg.ServiceName)
	router := gin.New()
	router.Use(gin.Recovery(), middleware.WithRequestContextGin(), metrics.MiddlewareGin(cfg.ServiceName))
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/readyz", func(c *gin.Context) {
		ctx, stop := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer stop()
		if err := commonmysql.HealthCheck(ctx, mysqlDB); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "dependency": "mysql"})
			return
		}
		if err := commonredis.HealthCheck(ctx, redisClient); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "dependency": "redis"})
			return
		}
		if err := commonkafka.HealthCheck(ctx, cfg.KafkaBrokers); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "dependency": "kafka"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	router.GET("/metrics", metrics.HandlerGin())
	inventoryhttp.NewHandler(svc).Register(router)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
	}

	logger.Info("inventory-service starting", "addr", server.Addr)
	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case err := <-serverErr:
		if err != nil {
			return err
		}
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig.String())
	}

	cancel()
	shutdownCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	return server.Shutdown(shutdownCtx)
}
