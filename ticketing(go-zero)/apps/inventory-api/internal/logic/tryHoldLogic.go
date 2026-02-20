// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"ticketing-gozero/apps/inventory-api/internal/svc"
	"ticketing-gozero/apps/inventory-api/internal/types"
	inventorysvc "ticketing-gozero/pkg/core/inventory"

	"github.com/zeromicro/go-zero/core/logx"
)

type TryHoldLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTryHoldLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TryHoldLogic {
	return &TryHoldLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TryHoldLogic) TryHold(req *types.TryHoldReq) (resp *types.PartitionStateResp, err error) {
	state, err := l.svcCtx.InventoryService.TryHold(l.ctx, inventorysvc.TryHoldInput{
		PartitionKey: req.PartitionKey,
		HoldID:       req.HoldID,
		Qty:          req.Qty,
		Capacity:     req.Capacity,
	})
	if err != nil {
		return nil, err
	}
	return &types.PartitionStateResp{
		PartitionKey: state.PartitionKey,
		Capacity:     state.Capacity,
		Available:    state.Available,
		Confirmed:    state.Confirmed,
		LastSeq:      state.LastSeq,
	}, nil
}
