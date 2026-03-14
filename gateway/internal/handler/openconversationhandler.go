package handler

import (
	"net/http"

	"pallink/gateway/internal/logic"
	"pallink/gateway/internal/svc"
	"pallink/gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func OpenConversationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OpenConversationReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewOpenConversationLogic(r.Context(), svcCtx)
		resp, err := l.OpenConversation(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
