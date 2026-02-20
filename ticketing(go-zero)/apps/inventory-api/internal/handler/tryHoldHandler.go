// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"ticketing-gozero/apps/inventory-api/internal/logic"
	"ticketing-gozero/apps/inventory-api/internal/svc"
	"ticketing-gozero/apps/inventory-api/internal/types"
)

func TryHoldHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TryHoldReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewTryHoldLogic(r.Context(), svcCtx)
		resp, err := l.TryHold(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
