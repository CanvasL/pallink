// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"pallink/gateway/internal/logic"
	"pallink/gateway/internal/svc"
	"pallink/gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetNotificationsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetNotificationsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetNotificationsLogic(r.Context(), svcCtx)
		resp, err := l.GetNotifications(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
