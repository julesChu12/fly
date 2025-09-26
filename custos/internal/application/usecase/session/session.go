package session

import (
	"context"
	"time"

	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"github.com/julesChu12/fly/custos/internal/domain/service/token"
	"github.com/julesChu12/fly/custos/pkg/errors"
)

type SessionUseCase struct {
	userRepo     repository.UserRepository
	tokenService *token.TokenService
	// Future: session repository will be added here
}

func NewSessionUseCase(userRepo repository.UserRepository, tokenService *token.TokenService) *SessionUseCase {
	return &SessionUseCase{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	UserID    uint      `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
}

// CreateSession creates a new user session
func (uc *SessionUseCase) CreateSession(ctx context.Context, userID uint, ttl time.Duration) (*Session, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewUserNotFoundError()
	}

	// Generate token for the session
	sessionID := uc.tokenService.GenerateSessionID()
	tokenPair, err := uc.tokenService.GenerateAccessToken(sessionID, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Token:     tokenPair.AccessToken,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	// TODO: Store session in repository
	// For now, just return the session

	return session, nil
}

// ValidateSession validates a session token
func (uc *SessionUseCase) ValidateSession(ctx context.Context, tokenString string) (*Session, error) {
	claims, err := uc.tokenService.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// TODO: Check session in repository
	// For now, create a session from token claims
	var expiresAt, createdAt time.Time
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}
	if claims.IssuedAt != nil {
		createdAt = claims.IssuedAt.Time
	}
	
	session := &Session{
		ID:        uc.tokenService.GenerateSessionID(),
		UserID:    claims.UserID,
		Token:     tokenString,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	return session, nil
}

// RefreshSession refreshes an existing session
func (uc *SessionUseCase) RefreshSession(ctx context.Context, sessionID string, ttl time.Duration) (*Session, error) {
	// TODO: Get session from repository
	// For now, return error
	return nil, errors.NewSessionNotFoundError()
}

// RevokeSession revokes a session
func (uc *SessionUseCase) RevokeSession(ctx context.Context, sessionID string) error {
	// TODO: Mark session as inactive in repository
	// For now, return success
	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (uc *SessionUseCase) RevokeAllUserSessions(ctx context.Context, userID uint) error {
	// TODO: Mark all user sessions as inactive in repository
	// For now, return success
	return nil
}

// ListUserSessions lists all active sessions for a user
func (uc *SessionUseCase) ListUserSessions(ctx context.Context, userID uint) ([]*Session, error) {
	// TODO: Get all active sessions for user from repository
	// For now, return empty list
	return []*Session{}, nil
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	// TODO: Implement proper session ID generation
	return "session_" + time.Now().Format("20060102150405")
}
