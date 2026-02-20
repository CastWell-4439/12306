// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"fmt"

	"ticketing-gozero/apps/inventory-api/internal/svc"
	"ticketing-gozero/apps/inventory-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AvailabilityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAvailabilityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AvailabilityLogic {
	return &AvailabilityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AvailabilityLogic) Availability(req *types.AvailabilityReq) (resp *types.AvailabilityResp, err error) {
	available, ok, err := l.svcCtx.InventoryService.GetAvailability(l.ctx, req.PartitionKey)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("partition not found")
	}
	return &types.AvailabilityResp{
		PartitionKey: req.PartitionKey,
		Available:    available,
	}, nil
}
