// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"ticketing-gozero/apps/order-api/internal/svc"
	"ticketing-gozero/apps/order-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderLogic) GetOrder(req *types.GetOrderReq) (resp *types.OrderResp, err error) {
	order, err := l.svcCtx.OrderService.GetOrder(l.ctx, req.OrderID)
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
