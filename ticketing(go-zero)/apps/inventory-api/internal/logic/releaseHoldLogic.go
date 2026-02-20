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

type ReleaseHoldLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReleaseHoldLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseHoldLogic {
	return &ReleaseHoldLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReleaseHoldLogic) ReleaseHold(req *types.ReleaseHoldReq) (resp *types.PartitionStateResp, err error) {
	state, err := l.svcCtx.InventoryService.ReleaseHold(l.ctx, inventorysvc.ReleaseInput{
		PartitionKey: req.PartitionKey,
		HoldID:       req.HoldID,
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
