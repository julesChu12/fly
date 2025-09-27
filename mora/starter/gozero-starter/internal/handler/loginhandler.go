package handler

import (
	"net/http"
	"time"

	"github.com/julesChu12/fly/mora/pkg/auth"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/julesChu12/fly/mora/starter/gozero-starter/internal/svc"
	"github.com/julesChu12/fly/mora/starter/gozero-starter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Example of using logger with trace context
		logger.WithCtx(r.Context()).Info("user login attempt")

		var req types.LoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			logger.WithCtx(r.Context()).Error("invalid login request", "error", err.Error())
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// Mock authentication - in production, validate against UserService
		if req.Username == "admin" && req.Password == "password" {
			// Generate access token using Mora auth
			tokenTTL := time.Duration(svcCtx.Config.JWT.TTL) * time.Second
			token, err := auth.GenerateToken("user-123", req.Username, svcCtx.Config.JWT.Secret, tokenTTL)
			if err != nil {
				logger.WithCtx(r.Context()).Error("token generation failed", "error", err.Error())
				httpx.Error(w, err)
				return
			}

			resp := &types.LoginResponse{
				AccessToken: token,
				TokenType:   "Bearer",
				ExpiresIn:   int(tokenTTL.Seconds()),
				UserID:      "user-123",
				Username:    req.Username,
			}

			logger.WithCtx(r.Context()).Info("user login success", "user_id", "user-123", "username", req.Username)
			httpx.OkJson(w, resp)
			return
		}

		// Authentication failed
		logger.WithCtx(r.Context()).Warn("authentication failed", "username", req.Username)
		httpx.WriteJson(w, http.StatusUnauthorized, map[string]string{
			"error":   "authentication failed",
			"message": "invalid username or password",
		})
	}
}
