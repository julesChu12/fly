package session

import (
	"context"
	"fmt"
	"time"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"github.com/julesChu12/fly/custos/internal/domain/service/token"
	"github.com/julesChu12/fly/custos/pkg/errors"
)

type SessionUseCase struct {
	userRepo     repository.UserRepository
	sessionRepo  repository.SessionRepository
	tokenService *token.TokenService
}

func NewSessionUseCase(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, tokenService *token.TokenService) *SessionUseCase {
	return &SessionUseCase{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		tokenService: tokenService,
	}
}

// CreateSession creates a new user session
func (uc *SessionUseCase) CreateSession(ctx context.Context, userID uint, userAgent, ip string) (*entity.Session, error) {
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewUserNotFoundError()
	}

	// Create session entity
	session := entity.NewSession(userID, userAgent, ip)

	// Store session in repository
	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// ValidateSession validates a session by ID
func (uc *SessionUseCase) ValidateSession(ctx context.Context, sessionID string) (*entity.Session, error) {
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if !session.IsValid() {
		return nil, fmt.Errorf("session has been revoked")
	}

	// Update last seen
	session.UpdateLastSeen()
	if err := uc.sessionRepo.UpdateLastSeen(ctx, sessionID, session.LastSeenAt); err != nil {
		// Log error but don't fail the validation
		// This is not critical for session validation
	}

	return session, nil
}

// RevokeSession revokes a session
func (uc *SessionUseCase) RevokeSession(ctx context.Context, sessionID string) error {
	now := time.Now()
	return uc.sessionRepo.Revoke(ctx, sessionID, now)
}

// RevokeAllUserSessions revokes all sessions for a user
func (uc *SessionUseCase) RevokeAllUserSessions(ctx context.Context, userID uint) error {
	now := time.Now()
	return uc.sessionRepo.RevokeByUser(ctx, userID, now)
}

// ListUserSessions lists all active sessions for a user
func (uc *SessionUseCase) ListUserSessions(ctx context.Context, userID uint) ([]*entity.Session, error) {
	now := time.Now()
	return uc.sessionRepo.ListActiveByUser(ctx, userID, now)
}

// CleanupExpiredSessions removes expired sessions
func (uc *SessionUseCase) CleanupExpiredSessions(ctx context.Context) error {
	// Clean up sessions older than 30 days
	cutoff := time.Now().AddDate(0, 0, -30)
	return uc.sessionRepo.CleanupExpired(ctx, cutoff)
}
