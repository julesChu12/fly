package auth

import (
	"context"

	"github.com/julesChu12/fly/custos/internal/application/dto"
	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"github.com/julesChu12/fly/custos/internal/domain/service/auth"
)

type RegisterUseCase struct {
	authService     *auth.AuthService
	userProfileRepo repository.UserProfileRepository
}

func NewRegisterUseCase(authService *auth.AuthService, userProfileRepo repository.UserProfileRepository) *RegisterUseCase {
	return &RegisterUseCase{
		authService:     authService,
		userProfileRepo: userProfileRepo,
	}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, req *dto.RegisterRequest) (*dto.UserInfo, error) {
	user, err := uc.authService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return uc.buildUserInfo(ctx, user)
}

type LoginUseCase struct {
	authService     *auth.AuthService
	userProfileRepo repository.UserProfileRepository
}

func NewLoginUseCase(authService *auth.AuthService, userProfileRepo repository.UserProfileRepository) *LoginUseCase {
	return &LoginUseCase{
		authService:     authService,
		userProfileRepo: userProfileRepo,
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

	userInfo, err := uc.buildUserInfo(ctx, user)
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
		User:             userInfo,
	}, nil
}

type RefreshUseCase struct {
	authService     *auth.AuthService
	userProfileRepo repository.UserProfileRepository
}

func NewRefreshUseCase(authService *auth.AuthService, userProfileRepo repository.UserProfileRepository) *RefreshUseCase {
	return &RefreshUseCase{
		authService:     authService,
		userProfileRepo: userProfileRepo,
	}
}

func (uc *RefreshUseCase) Execute(ctx context.Context, req *dto.RefreshRequest) (*dto.LoginResponse, error) {
	tokenPair, user, err := uc.authService.Refresh(ctx, req.SessionID, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	userInfo, err := uc.buildUserInfo(ctx, user)
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
		User:             userInfo,
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

// buildUserInfo builds UserInfo DTO from user entity and profile
func (uc *RegisterUseCase) buildUserInfo(ctx context.Context, user *entity.User) (*dto.UserInfo, error) {
	userInfo := &dto.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
		Status:   string(user.Status),
	}

	// Fetch profile data
	profile, err := uc.userProfileRepo.GetByUserID(ctx, user.ID)
	if err == nil {
		userInfo.Nickname = profile.Nickname
		userInfo.Avatar = profile.Avatar
	}
	// Ignore profile not found error - nickname/avatar will be empty

	return userInfo, nil
}

// buildUserInfo builds UserInfo DTO from user entity and profile
func (uc *LoginUseCase) buildUserInfo(ctx context.Context, user *entity.User) (*dto.UserInfo, error) {
	userInfo := &dto.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
		Status:   string(user.Status),
	}

	// Fetch profile data
	profile, err := uc.userProfileRepo.GetByUserID(ctx, user.ID)
	if err == nil {
		userInfo.Nickname = profile.Nickname
		userInfo.Avatar = profile.Avatar
	}
	// Ignore profile not found error - nickname/avatar will be empty

	return userInfo, nil
}

// buildUserInfo builds UserInfo DTO from user entity and profile
func (uc *RefreshUseCase) buildUserInfo(ctx context.Context, user *entity.User) (*dto.UserInfo, error) {
	userInfo := &dto.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
		Status:   string(user.Status),
	}

	// Fetch profile data
	profile, err := uc.userProfileRepo.GetByUserID(ctx, user.ID)
	if err == nil {
		userInfo.Nickname = profile.Nickname
		userInfo.Avatar = profile.Avatar
	}
	// Ignore profile not found error - nickname/avatar will be empty

	return userInfo, nil
}
