package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"

	"ticketing-gozero/apps/ticket-worker/internal/svc"
	commonkafka "ticketing-gozero/pkg/infra/kafka"
	commonmysql "ticketing-gozero/pkg/infra/mysql"
	commonredis "ticketing-gozero/pkg/infra/redis"
	"ticketing-gozero/pkg/observability"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext, metrics *observability.Metrics) {
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/healthz", Handler: func(w http.ResponseWriter, r *http.Request) {
			httpx.OkJsonCtx(r.Context(), w, map[string]string{"status": "ok"})
		}},
		{Method: http.MethodGet, Path: "/readyz", Handler: func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := commonmysql.HealthCheck(ctx, serverCtx.MysqlDB); err != nil {
				httpx.WriteJson(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "dependency": "mysql"})
				return
			}
			if err := commonredis.HealthCheck(ctx, serverCtx.Redis); err != nil {
				httpx.WriteJson(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "dependency": "redis"})
				return
			}
			if err := commonkafka.HealthCheck(ctx, serverCtx.Config.Kafka.Brokers); err != nil {
				httpx.WriteJson(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "dependency": "kafka"})
				return
			}
			httpx.OkJsonCtx(r.Context(), w, map[string]string{"status": "ready"})
		}},
		{Method: http.MethodGet, Path: "/metrics", Handler: metrics.Handler()},
	})
}


