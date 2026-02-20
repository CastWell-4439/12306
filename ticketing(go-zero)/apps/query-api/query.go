// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"ticketing-gozero/apps/query-api/internal/config"
	"ticketing-gozero/apps/query-api/internal/handler"
	"ticketing-gozero/apps/query-api/internal/svc"
	querysvc "ticketing-gozero/pkg/core/query"
	commonkafka "ticketing-gozero/pkg/infra/kafka"
	"ticketing-gozero/pkg/infra/logging"
	commonmysql "ticketing-gozero/pkg/infra/mysql"
	querycache "ticketing-gozero/pkg/infra/query/cache"
	queryreadmodel "ticketing-gozero/pkg/infra/query/readmodel"
	commonredis "ticketing-gozero/pkg/infra/redis"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/query-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

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

	if err := commonkafka.HealthCheck(context.Background(), c.Kafka.Brokers); err != nil {
		panic(fmt.Sprintf("kafka: %v", err))
	}
	orderConsumer := commonkafka.NewConsumer(c.Kafka.Brokers, "order.events", "query-order-consumer")
	defer orderConsumer.Close()
	ticketConsumer := commonkafka.NewConsumer(c.Kafka.Brokers, "ticket.events", "query-ticket-consumer")
	defer ticketConsumer.Close()

	queryService := querysvc.NewService(
		logger,
		queryreadmodel.NewRepository(mysqlDB),
		querycache.NewStore(redisClient, 30*time.Second),
	)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = queryService.RebuildColdStart(rootCtx)
	go queryService.StartOrderEventsConsumer(rootCtx, orderConsumer)
	go queryService.StartTicketEventsConsumer(rootCtx, ticketConsumer)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	ctx.QueryService = queryService
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting query-api at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
