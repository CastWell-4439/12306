package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	commonconfig "ticketing/internal/common/config"
	commonkafka "ticketing/internal/common/kafka"
	"ticketing/internal/common/logging"
	commonmetrics "ticketing/internal/common/metrics"
	"ticketing/internal/common/middleware"
	commonmysql "ticketing/internal/common/mysql"
	commonredis "ticketing/internal/common/redis"
	"ticketing/internal/ticket/application"
	grpcclient "ticketing/internal/ticket/infrastructure/grpc_client"
	"ticketing/internal/ticket/infrastructure/outbox"
	"ticketing/internal/ticket/infrastructure/repository"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("ticket-worker stopped with error: %v", err)
	}
}

func run() error {
	cfg := commonconfig.Load("ticket-worker")
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

	producer := commonkafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()
	if err := commonkafka.HealthCheck(context.Background(), cfg.KafkaBrokers); err != nil {
		return fmt.Errorf("kafka init failed: %w", err)
	}
	consumer := commonkafka.NewConsumer(cfg.KafkaBrokers, "order.events", "ticket-worker")
	defer consumer.Close()

	seatAllocator := grpcclient.SeatAllocatorClient(grpcclient.NewMockSeatAllocator())
	mode := strings.ToLower(cfg.SeatAllocatorMode)
	if mode == "grpc" {
		client, err := grpcclient.NewGRPCSeatAllocator(
			cfg.SeatAllocatorAddr,
			cfg.SeatAllocatorTrainID,
			cfg.SeatAllocatorTravelDate,
			cfg.SeatAllocatorCoachType,
			cfg.SeatAllocatorFromIndex,
			cfg.SeatAllocatorToIndex,
		)
		if err != nil {
			return fmt.Errorf("init grpc seat allocator failed: %w", err)
		}
		seatAllocator = client
		defer client.Close()
	}
	logger.Info("seat allocator selected", "mode", mode, "addr", cfg.SeatAllocatorAddr)

	worker := application.NewWorker(
		logger,
		consumer,
		producer,
		repository.NewRepository(mysqlDB),
		outbox.NewRepository(mysqlDB),
		seatAllocator,
	)
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := worker.Start(rootCtx); err != nil {
			logger.Error("ticket worker stopped", "error", err)
		}
	}()

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

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
	}
	logger.Info("ticket-worker starting", "addr", server.Addr)

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
