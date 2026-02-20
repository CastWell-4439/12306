// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"ticketing-gozero/apps/query-api/internal/config"
	querysvc "ticketing-gozero/pkg/core/query"
)

type ServiceContext struct {
	Config       config.Config
	QueryService *querysvc.Service
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
