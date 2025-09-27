package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"github.com/julesChu12/fly/custos/internal/domain/service/token"
	"github.com/julesChu12/fly/custos/pkg/constants"
	"github.com/julesChu12/fly/custos/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo         repository.UserRepository
	sessionRepo      repository.SessionRepository
	refreshTokenRepo repository.RefreshTokenRepository
	tokenService     *token.TokenService
}

func NewAuthService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, refreshTokenRepo repository.RefreshTokenRepository, tokenService *token.TokenService) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		sessionRepo:      sessionRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenService:     tokenService,
	}
}

type LoginMetadata struct {
	IPAddress string
	UserAgent string
}

func (s *AuthService) Register(ctx context.Context, username, email, password string) (*entity.User, error) {
	if len(username) < constants.UsernameMinLength || len(username) > constants.UsernameMaxLength {
		return nil, errors.NewInvalidPasswordError(
			fmt.Sprintf("Username must be between %d and %d characters",
				constants.UsernameMinLength, constants.UsernameMaxLength))
	}

	if len(password) < constants.PasswordMinLength || len(password) > constants.PasswordMaxLength {
		return nil, errors.NewInvalidPasswordError(
			fmt.Sprintf("Password must be between %d and %d characters",
				constants.PasswordMinLength, constants.PasswordMaxLength))
	}

	exists, err := s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if exists {
		return nil, errors.NewUserAlreadyExistsError(username)
	}

	exists, err = s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, errors.NewUserAlreadyExistsError(email)
	}

	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := entity.NewUser(username, email, hashedPassword)
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string, meta *LoginMetadata) (*token.TokenPair, *entity.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, nil, errors.NewInvalidCredentialsError()
	}

	if !user.IsActive() {
		return nil, nil, errors.NewInvalidCredentialsError()
	}

	if !s.checkPassword(password, user.Password) {
		return nil, nil, errors.NewInvalidCredentialsError()
	}

	// Create session entity first to get the session ID
	session := entity.NewSession(user.ID, "", "")
	if meta != nil {
		session.UserAgent = meta.UserAgent
		session.IP = meta.IPAddress
	}

	// Generate tokens using the session ID from the entity
	tokenPair, err := s.tokenService.GenerateAccessToken(session.SessionID, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create refresh token entity first
	refreshTokenEntity := entity.NewRefreshToken(user.ID, refreshToken.Token, refreshToken.ExpiresAt)
	if err := s.refreshTokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	// Associate refresh token with session
	session.RefreshTokenID = &refreshTokenEntity.ID

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		// If session creation fails, clean up the refresh token
		_ = s.refreshTokenRepo.Delete(ctx, refreshTokenEntity.ID)
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	tokenPair.RefreshToken = refreshToken.Token
	tokenPair.RefreshExpiresIn = refreshToken.ExpiresIn
	tokenPair.SessionID = session.SessionID

	return tokenPair, user, nil
}

func (s *AuthService) Refresh(ctx context.Context, sessionID, refreshToken string) (*token.TokenPair, *entity.User, error) {
	// Validate refresh token by getting session associated with it
	hashedRefreshToken := s.tokenService.HashRefreshToken(refreshToken)
	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, hashedRefreshToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}
	if session == nil {
		return nil, nil, errors.NewTokenInvalidError()
	}

	// Verify the session ID matches
	if session.SessionID != sessionID {
		return nil, nil, errors.NewTokenInvalidError()
	}

	now := time.Now()
	if !session.IsValid() {
		_ = s.sessionRepo.Revoke(ctx, sessionID, now)
		return nil, nil, errors.NewTokenExpiredError()
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, errors.NewUserNotFoundError()
	}
	if !user.IsActive() {
		_ = s.sessionRepo.Revoke(ctx, sessionID, now)
		return nil, nil, errors.NewInvalidCredentialsError()
	}

	// Generate new access token
	tokenPair, err := s.tokenService.GenerateAccessToken(session.SessionID, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Generate new refresh token for rotation
	newRefresh, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update session with new refresh token (this will mark old token as used)
	if err := s.sessionRepo.UpdateRefreshToken(ctx, session.SessionID, s.tokenService.HashRefreshToken(newRefresh.Token), newRefresh.ExpiresAt, now); err != nil {
		return nil, nil, fmt.Errorf("failed to rotate refresh token: %w", err)
	}

	tokenPair.RefreshToken = newRefresh.Token
	tokenPair.RefreshExpiresIn = newRefresh.ExpiresIn
	tokenPair.SessionID = session.SessionID

	return tokenPair, user, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.NewSessionNotFoundError()
	}
	if err := s.sessionRepo.Revoke(ctx, sessionID, time.Now()); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uint) error {
	if userID == 0 {
		return errors.NewUserNotFoundError()
	}
	if err := s.sessionRepo.RevokeByUser(ctx, userID, time.Now()); err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}
	return nil
}

func (s *AuthService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *AuthService) checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
