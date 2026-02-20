// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"ticketing-gozero/apps/query-api/internal/svc"
	"ticketing-gozero/apps/query-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderViewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderViewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderViewLogic {
	return &GetOrderViewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderViewLogic) GetOrderView(req *types.GetOrderViewReq) (resp *types.OrderViewResp, err error) {
	v, err := l.svcCtx.QueryService.GetOrderView(l.ctx, req.OrderID)
	if err != nil {
		return nil, err
	}
	return &types.OrderViewResp{
		OrderID:       v.OrderID,
		Status:        v.Status,
		AmountCents:   v.AmountCents,
		ProviderTxnID: v.ProviderTxnID,
		SeatNo:        v.SeatNo,
		UpdatedAt:     v.UpdatedAt.String(),
	}, nil
}
