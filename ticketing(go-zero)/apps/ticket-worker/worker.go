package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	"ticketing-gozero/apps/ticket-worker/internal/config"
	"ticketing-gozero/apps/ticket-worker/internal/handler"
	"ticketing-gozero/apps/ticket-worker/internal/svc"
	ticketsvc "ticketing-gozero/pkg/core/ticket"
	commonkafka "ticketing-gozero/pkg/infra/kafka"
	"ticketing-gozero/pkg/infra/logging"
	"ticketing-gozero/pkg/infra/middleware"
	commonmysql "ticketing-gozero/pkg/infra/mysql"
	commonredis "ticketing-gozero/pkg/infra/redis"
	grpcclient "ticketing-gozero/pkg/infra/ticket/grpc_client"
	ticketoutbox "ticketing-gozero/pkg/infra/ticket/outbox"
	ticketrepo "ticketing-gozero/pkg/infra/ticket/repository"
	"ticketing-gozero/pkg/observability"
)

var configFile = flag.String("f", "etc/ticket-worker.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	logger := logging.New(c.Name, "dev", "gozero")

	mysqlDB, err := commonmysql.New(c.Mysql.Dsn)
	if err != nil {
		panic(fmt.Sprintf("mysql init failed: %v", err))
	}
	defer mysqlDB.Close()

	redisClient, err := commonredis.New(c.Redis.Addr, c.Redis.Password, c.Redis.DB)
	if err != nil {
		panic(fmt.Sprintf("redis init failed: %v", err))
	}
	defer redisClient.Close()

	producer := commonkafka.NewProducer(c.Kafka.Brokers)
	defer producer.Close()
	if err := commonkafka.HealthCheck(context.Background(), c.Kafka.Brokers); err != nil {
		panic(fmt.Sprintf("kafka init failed: %v", err))
	}
	consumer := commonkafka.NewConsumer(c.Kafka.Brokers, "order.events", "ticket-worker")
	defer consumer.Close()

	seatAllocator := grpcclient.SeatAllocatorClient(grpcclient.NewMockSeatAllocator())
	mode := strings.ToLower(c.SeatAllocatorMode)
	if mode == "grpc" {
		client, err := grpcclient.NewGRPCSeatAllocator(
			c.SeatAllocatorAddr,
			c.SeatAllocatorTrainID,
			c.SeatAllocatorTravelDate,
			c.SeatAllocatorCoachType,
			c.SeatAllocatorFromIndex,
			c.SeatAllocatorToIndex,
		)
		if err != nil {
			panic(fmt.Sprintf("init grpc seat allocator failed: %v", err))
		}
		seatAllocator = client
		defer client.Close()
	}
	logger.Info("seat allocator selected", "mode", mode, "addr", c.SeatAllocatorAddr)

	worker := ticketsvc.NewWorker(
		logger,
		consumer,
		producer,
		ticketrepo.NewRepository(mysqlDB),
		ticketoutbox.NewRepository(mysqlDB),
		seatAllocator,
	)
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := worker.Start(rootCtx); err != nil {
			logger.Error("ticket worker stopped", "error", err)
		}
	}()

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	metrics := observability.New(c.Name)
	server.Use(middleware.WithRequestContext)
	server.Use(metrics.Middleware(c.Name))

	serverCtx := svc.NewServiceContext(c, mysqlDB, redisClient)
	handler.RegisterHandlers(server, serverCtx, metrics)

	fmt.Printf("Starting ticket-worker at %s:%d...\n", c.Host, c.Port)
	server.Start()
}


