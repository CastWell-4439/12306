// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"

	"ticketing-gozero/apps/gateway-api/internal/config"
	"ticketing-gozero/apps/gateway-api/internal/handler"
	"ticketing-gozero/apps/gateway-api/internal/svc"
	commonmysql "ticketing-gozero/pkg/infra/mysql"
	commonredis "ticketing-gozero/pkg/infra/redis"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/gateway-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

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

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	ctx.MysqlDB = mysqlDB
	ctx.Redis = redisClient
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting gateway-api at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
