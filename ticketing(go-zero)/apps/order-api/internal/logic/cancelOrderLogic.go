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

type CancelOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelOrderLogic) CancelOrder(req *types.CancelOrderReq) (resp *types.OrderResp, err error) {
	order, err := l.svcCtx.OrderService.CancelOrder(l.ctx, ordersvc.CancelOrderInput{
		OrderID:      req.OrderID,
		PartitionKey: req.PartitionKey,
		HoldID:       req.HoldID,
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
