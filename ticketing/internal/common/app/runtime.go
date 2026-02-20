package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	redisv9 "github.com/redis/go-redis/v9"

	"ticketing/internal/common/config"
	commonerrors "ticketing/internal/common/errors"
	commonkafka "ticketing/internal/common/kafka"
	"ticketing/internal/common/logging"
	commonmetrics "ticketing/internal/common/metrics"
	"ticketing/internal/common/middleware"
	commonmysql "ticketing/internal/common/mysql"
	commonredis "ticketing/internal/common/redis"
)

type Runtime struct {
	cfg        config.Config
	logger     *slog.Logger
	mysqlDB    *sql.DB
	redis      *redisv9.Client
	kafkaWrite *commonkafka.Producer
}

func Run(serviceName string) error {
	rt, err := newRuntime(serviceName)
	if err != nil {
		return err
	}
	defer rt.close()

	router := gin.New()
	metrics := commonmetrics.New(rt.cfg.ServiceName)
	router.Use(gin.Recovery(), middleware.WithRequestContextGin(), metrics.MiddlewareGin(rt.cfg.ServiceName))
	router.GET("/healthz", rt.healthHandler)
	router.GET("/readyz", rt.readyHandler)
	router.GET("/metrics", metrics.HandlerGin())

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", rt.cfg.HTTPPort),
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
	}

	rt.logger.Info("service starting", "addr", server.Addr)
	serverErrCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
		close(serverErrCh)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case err := <-serverErrCh:
		if err != nil {
			return err
		}
	case sig := <-sigCh:
		rt.logger.Info("shutdown signal received", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

func newRuntime(serviceName string) (*Runtime, error) {
	cfg := config.Load(serviceName)
	logger := logging.New(cfg.ServiceName, cfg.Env, cfg.Version)

	mysqlDB, err := commonmysql.New(cfg.MySQLDSN)
	if err != nil {
		return nil, fmt.Errorf("mysql init failed: %w", err)
	}

	redisClient, err := commonredis.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		_ = mysqlDB.Close()
		return nil, fmt.Errorf("redis init failed: %w", err)
	}

	kafkaWriter := commonkafka.NewProducer(cfg.KafkaBrokers)
	if err := commonkafka.HealthCheck(context.Background(), cfg.KafkaBrokers); err != nil {
		_ = kafkaWriter.Close()
		_ = redisClient.Close()
		_ = mysqlDB.Close()
		return nil, fmt.Errorf("kafka init failed: %w", err)
	}

	return &Runtime{
		cfg:        cfg,
		logger:     logger,
		mysqlDB:    mysqlDB,
		redis:      redisClient,
		kafkaWrite: kafkaWriter,
	}, nil
}

func (r *Runtime) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (r *Runtime) readyHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.DefaultRequestTimeout())
	defer cancel()

	if err := commonmysql.HealthCheck(ctx, r.mysqlDB); err != nil {
		r.failReady(c, "mysql", err)
		return
	}
	if err := commonredis.HealthCheck(ctx, r.redis); err != nil {
		r.failReady(c, "redis", err)
		return
	}
	if err := commonkafka.HealthCheck(ctx, r.cfg.KafkaBrokers); err != nil {
		r.failReady(c, "kafka", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (r *Runtime) failReady(c *gin.Context, dep string, err error) {
	r.logger.Error("readiness check failed",
		"error", err,
		"dependency", dep,
		"kind", commonerrors.ErrDependencyUnavailable.Error(),
	)
	c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "dependency": dep})
}

func (r *Runtime) close() {
	if r.kafkaWrite != nil {
		_ = r.kafkaWrite.Close()
	}
	if r.redis != nil {
		_ = r.redis.Close()
	}
	if r.mysqlDB != nil {
		_ = r.mysqlDB.Close()
	}
}
