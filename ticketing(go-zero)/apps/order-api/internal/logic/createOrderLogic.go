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

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (resp *types.OrderResp, err error) {
	order, err := l.svcCtx.OrderService.CreateOrder(l.ctx, ordersvc.CreateOrderInput{
		IdempotencyKey: req.IdempotencyKey,
		AmountCents:    req.AmountCents,
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
