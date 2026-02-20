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

type PaymentCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPaymentCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PaymentCallbackLogic {
	return &PaymentCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PaymentCallbackLogic) PaymentCallback(req *types.PaymentCallbackReq) (resp *types.OrderResp, err error) {
	order, err := l.svcCtx.OrderService.PaymentCallback(l.ctx, ordersvc.PaymentCallbackInput{
		OrderID:       req.OrderID,
		ProviderTxnID: req.ProviderTxnID,
		Status:        req.Status,
		PartitionKey:  req.PartitionKey,
		HoldID:        req.HoldID,
		Signature:     req.Signature,
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
