// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"database/sql"

	redisv9 "github.com/redis/go-redis/v9"

	"ticketing-gozero/apps/gateway-api/internal/config"
)

type ServiceContext struct {
	Config  config.Config
	MysqlDB *sql.DB
	Redis   *redisv9.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
// goctl 1.9.2

package svc

import (
	"ticketing-gozero/apps/gateway-api/internal/config"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}

