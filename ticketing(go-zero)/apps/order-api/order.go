// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"context"
	"flag"
	"fmt"

	"ticketing-gozero/apps/order-api/internal/config"
	"ticketing-gozero/apps/order-api/internal/handler"
	"ticketing-gozero/apps/order-api/internal/svc"
	ordersvc "ticketing-gozero/pkg/core/order"
	commonkafka "ticketing-gozero/pkg/infra/kafka"
	"ticketing-gozero/pkg/infra/logging"
	commonmysql "ticketing-gozero/pkg/infra/mysql"
	orderevent "ticketing-gozero/pkg/infra/order/event"
	orderinventory "ticketing-gozero/pkg/infra/order/inventory"
	orderoutbox "ticketing-gozero/pkg/infra/order/outbox"
	orderrepo "ticketing-gozero/pkg/infra/order/repository"
	commonredis "ticketing-gozero/pkg/infra/redis"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/order-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// ---- infrastructure init (keep original reliability) ----
	logger := logging.New(c.Name, "prod", "gozero")
	mysqlDB, err := commonmysql.New(c.Mysql.Dsn)
	if err != nil {
		panic(fmt.Sprintf("mysql: %v", err))
	}
	defer mysqlDB.Close()

	redisClient, err := commonredis.New(c.Redis.Addr, c.Redis.Password, c.Redis.DB)
	if err != nil {
		panic(fmt.Sprintf("redis: %v", err))
	}
	defer redisClient.Close()

	kafkaProducer := commonkafka.NewProducer(c.Kafka.Brokers)
	defer kafkaProducer.Close()

	// ---- core service wiring ----
	orderService := ordersvc.NewService(
		logger,
		orderrepo.NewRepository(mysqlDB),
		orderoutbox.NewRepository(mysqlDB),
		orderevent.NewPublisher(kafkaProducer, "order.events"),
		orderinventory.NewClient(c.InventoryServiceURL),
		ordersvc.Config{
			DefaultPartitionKey: c.OrderInventoryPartitionKey,
			DefaultHoldQty:      c.OrderInventoryDefaultQty,
			DefaultCapacity:     c.OrderInventoryCapacity,
			PaymentSignKey:      c.PaymentCallbackSignKey,
		},
	)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go orderService.StartOutboxPublisher(rootCtx)

	// ---- go-zero server (带内置中间件) ----
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	ctx.OrderService = orderService
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting order-api at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
