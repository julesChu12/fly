package handler

import (
	"net/http"
	"time"

	gozeroauth "github.com/julesChu12/fly/mora/adapters/gozero"
	"github.com/julesChu12/fly/mora/starter/gozero-starter/internal/svc"
	"github.com/julesChu12/fly/mora/starter/gozero-starter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ProtectedHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := gozeroauth.GetUserID(r.Context())

		resp := &types.ProtectedResponse{
			Message: "This is a protected endpoint",
			UserID:  userID,
			Time:    time.Now().Format(time.RFC3339),
		}

		httpx.OkJson(w, resp)
	}
}
