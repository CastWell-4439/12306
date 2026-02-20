// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"ticketing-gozero/apps/inventory-api/internal/config"
	inventorysvc "ticketing-gozero/pkg/core/inventory"
)

type ServiceContext struct {
	Config           config.Config
	InventoryService *inventorysvc.Service
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
