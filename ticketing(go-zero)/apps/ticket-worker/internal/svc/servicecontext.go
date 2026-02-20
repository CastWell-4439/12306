package svc

import (
	"database/sql"

	redisv9 "github.com/redis/go-redis/v9"

	"ticketing-gozero/apps/ticket-worker/internal/config"
)

type ServiceContext struct {
	Config  config.Config
	MysqlDB *sql.DB
	Redis   *redisv9.Client
}

func NewServiceContext(c config.Config, db *sql.DB, redis *redisv9.Client) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		MysqlDB: db,
		Redis:   redis,
	}
}

