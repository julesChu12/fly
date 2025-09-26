package auth

import (
	"context"

	"github.com/julesChu12/fly/custos/internal/application/dto"
	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/service/auth"
)

type RegisterUseCase struct {
	authService *auth.AuthService
}

func NewRegisterUseCase(authService *auth.AuthService) *RegisterUseCase {
	return &RegisterUseCase{
		authService: authService,
	}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, req *dto.RegisterRequest) (*dto.UserInfo, error) {
	user, err := uc.authService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &dto.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Role:     string(user.Role),
		Status:   string(user.Status),
	}, nil
}

type LoginUseCase struct {
	authService *auth.AuthService
}

func NewLoginUseCase(authService *auth.AuthService) *LoginUseCase {
	return &LoginUseCase{
		authService: authService,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, req *dto.LoginRequest, meta *dto.LoginMetadata) (*dto.LoginResponse, error) {
	var domainMeta *auth.LoginMetadata
	if meta != nil {
		domainMeta = &auth.LoginMetadata{IPAddress: meta.IPAddress, UserAgent: meta.UserAgent}
	}

	tokenPair, user, err := uc.authService.Login(ctx, req.Username, req.Password, domainMeta)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:      tokenPair.AccessToken,
		TokenType:        tokenPair.TokenType,
		ExpiresIn:        tokenPair.ExpiresIn,
		RefreshToken:     tokenPair.RefreshToken,
		RefreshExpiresIn: tokenPair.RefreshExpiresIn,
		SessionID:        tokenPair.SessionID,
		User: &dto.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Role:     string(user.Role),
			Status:   string(user.Status),
		},
	}, nil
}

type RefreshUseCase struct {
	authService *auth.AuthService
}

func NewRefreshUseCase(authService *auth.AuthService) *RefreshUseCase {
	return &RefreshUseCase{authService: authService}
}

func (uc *RefreshUseCase) Execute(ctx context.Context, req *dto.RefreshRequest) (*dto.LoginResponse, error) {
	tokenPair, user, err := uc.authService.Refresh(ctx, req.SessionID, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:      tokenPair.AccessToken,
		TokenType:        tokenPair.TokenType,
		ExpiresIn:        tokenPair.ExpiresIn,
		RefreshToken:     tokenPair.RefreshToken,
		RefreshExpiresIn: tokenPair.RefreshExpiresIn,
		SessionID:        tokenPair.SessionID,
		User:             entityToUserInfo(user),
	}, nil
}

type LogoutUseCase struct {
	authService *auth.AuthService
}

func NewLogoutUseCase(authService *auth.AuthService) *LogoutUseCase {
	return &LogoutUseCase{authService: authService}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, sessionID string) error {
	return uc.authService.Logout(ctx, sessionID)
}

type LogoutAllUseCase struct {
	authService *auth.AuthService
}

func NewLogoutAllUseCase(authService *auth.AuthService) *LogoutAllUseCase {
	return &LogoutAllUseCase{authService: authService}
}

func (uc *LogoutAllUseCase) Execute(ctx context.Context, userID uint) error {
	return uc.authService.LogoutAll(ctx, userID)
}

func entityToUserInfo(user *entity.User) *dto.UserInfo {
	return &dto.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Role:     string(user.Role),
		Status:   string(user.Status),
	}
}
