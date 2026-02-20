// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"time"

	"ticketing-gozero/apps/gateway-api/internal/svc"
	"ticketing-gozero/apps/gateway-api/internal/types"
	commonkafka "ticketing-gozero/pkg/infra/kafka"
	commonmysql "ticketing-gozero/pkg/infra/mysql"
	commonredis "ticketing-gozero/pkg/infra/redis"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReadyzLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReadyzLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReadyzLogic {
	return &ReadyzLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReadyzLogic) Readyz() (resp *types.StatusResp, err error) {
	ctx, cancel := context.WithTimeout(l.ctx, 2*time.Second)
	defer cancel()

	if err := commonmysql.HealthCheck(ctx, l.svcCtx.MysqlDB); err != nil {
		return nil, err
	}
	if err := commonredis.HealthCheck(ctx, l.svcCtx.Redis); err != nil {
		return nil, err
	}
	if err := commonkafka.HealthCheck(ctx, l.svcCtx.Config.Kafka.Brokers); err != nil {
		return nil, err
	}
	return &types.StatusResp{Status: "ready"}, nil
}
