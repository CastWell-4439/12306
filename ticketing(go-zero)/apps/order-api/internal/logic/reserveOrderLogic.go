// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"ticketing-gozero/apps/order-api/internal/svc"
	"ticketing-gozero/apps/order-api/internal/types"
	ordersvc "ticketing-gozero/pkg/core/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReserveOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReserveOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReserveOrderLogic {
	return &ReserveOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReserveOrderLogic) ReserveOrder(req *types.ReserveOrderReq) (resp *types.OrderResp, err error) {
	order, err := l.svcCtx.OrderService.ReserveOrder(l.ctx, ordersvc.ReserveOrderInput{
		OrderID:      req.OrderID,
		PartitionKey: req.PartitionKey,
		HoldID:       req.HoldID,
		Qty:          req.Qty,
		Capacity:     req.Capacity,
	})
	if err != nil {
		return nil, err
	}
	return &types.OrderResp{
		OrderID:        order.OrderID,
		IdempotencyKey: order.IdempotencyKey,
		Status:         string(order.Status),
		AmountCents:    order.AmountCents,
		CreatedAt:      order.CreatedAt.String(),
		UpdatedAt:      order.UpdatedAt.String(),
	}, nil
}
