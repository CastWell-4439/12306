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
	"syscall"
	"time"

	commonconfig "ticketing/internal/common/config"
	commonkafka "ticketing/internal/common/kafka"
	"ticketing/internal/common/logging"
	commonmetrics "ticketing/internal/common/metrics"
	"ticketing/internal/common/middleware"
	commonmysql "ticketing/internal/common/mysql"
	commonredis "ticketing/internal/common/redis"
	"ticketing/internal/order/application"
	"ticketing/internal/order/infrastructure/event"
	inventoryclient "ticketing/internal/order/infrastructure/inventory"
	"ticketing/internal/order/infrastructure/outbox"
	"ticketing/internal/order/infrastructure/repository"
	orderhttp "ticketing/internal/order/interfaces/http"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("order-service stopped with error: %v", err)
	}
}

func run() error {
	cfg := commonconfig.Load("order-service")
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

	repo := repository.NewRepository(mysqlDB)
	outboxRepo := outbox.NewRepository(mysqlDB)
	publisher := event.NewPublisher(kafkaProducer, "order.events")
	inventoryAPI := inventoryclient.NewClient(cfg.InventoryServiceURL)
	svc := application.NewService(
		logger,
		repo,
		outboxRepo,
		publisher,
		inventoryAPI,
		application.Config{
			DefaultPartitionKey: cfg.OrderInventoryPartitionKey,
			DefaultHoldQty:      cfg.OrderInventoryDefaultQty,
			DefaultCapacity:     cfg.OrderInventoryCapacity,
			PaymentSignKey:      cfg.PaymentCallbackSignKey,
		},
	)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go svc.StartOutboxPublisher(rootCtx)

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
	orderhttp.NewHandler(svc).Register(router)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
	}
	logger.Info("order-service starting", "addr", server.Addr)

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
