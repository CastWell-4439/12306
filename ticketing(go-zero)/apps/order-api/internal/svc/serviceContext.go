// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"ticketing-gozero/apps/order-api/internal/config"
	ordersvc "ticketing-gozero/pkg/core/order"
)

type ServiceContext struct {
	Config       config.Config
	OrderService *ordersvc.Service
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
