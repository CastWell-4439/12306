// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"ticketing-gozero/apps/inventory-api/internal/config"
	"ticketing-gozero/apps/inventory-api/internal/handler"
	"ticketing-gozero/apps/inventory-api/internal/svc"
	inventorysvc "ticketing-gozero/pkg/core/inventory"
	commonkafka "ticketing-gozero/pkg/infra/kafka"
	"ticketing-gozero/pkg/infra/logging"
	commonmysql "ticketing-gozero/pkg/infra/mysql"
	commonredis "ticketing-gozero/pkg/infra/redis"
	invevent "ticketing-gozero/pkg/infra/inventory/event"
	invsnapshot "ticketing-gozero/pkg/infra/inventory/snapshot"
	invttl "ticketing-gozero/pkg/infra/inventory/ttl"
	invwal "ticketing-gozero/pkg/infra/inventory/wal"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/inventory-api.yaml", "the config file")

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

	kafkaProducer := commonkafka.NewProducer(c.Kafka.Brokers)
	defer kafkaProducer.Close()

	invService := inventorysvc.NewService(
		logger,
		invwal.NewRepository(mysqlDB),
		invsnapshot.NewRepository(mysqlDB),
		invevent.NewPublisher(kafkaProducer, "inventory.events"),
		invttl.NewStore(redisClient, time.Duration(c.InventoryHoldTTLSecs)*time.Second),
		inventorysvc.Config{
			ShardCount:           c.InventoryShardCount,
			WALBuffer:            c.InventoryWALBuffer,
			SnapshotInterval:     time.Duration(c.InventorySnapshotIntervalSecs) * time.Second,
			SnapshotOpsThreshold: c.InventorySnapshotOpsThreshold,
		},
	)
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := invService.Start(rootCtx); err != nil {
		panic(fmt.Sprintf("inventory start: %v", err))
	}

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	ctx.InventoryService = invService
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting inventory-api at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
