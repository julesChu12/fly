package auth

import (
	"context"
	stdErrors "errors"
	"testing"
	"time"

	"github.com/julesChu12/custos/internal/domain/entity"
	"github.com/julesChu12/custos/internal/domain/service/token"
	"github.com/julesChu12/custos/pkg/errors"
	"github.com/julesChu12/custos/pkg/types"
	"github.com/stretchr/testify/require"
)

type fakeUserRepo struct {
	byID       map[uint]*entity.User
	byUsername map[string]*entity.User
	byEmail    map[string]*entity.User
	nextID     uint
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		byID:       make(map[uint]*entity.User),
		byUsername: make(map[string]*entity.User),
		byEmail:    make(map[string]*entity.User),
		nextID:     1,
	}
}

type fakeSessionRepo struct {
	sessions map[string]*entity.Session
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{sessions: make(map[string]*entity.Session)}
}

func (r *fakeSessionRepo) Create(_ context.Context, session *entity.Session) error {
	clone := *session
	r.sessions[session.ID] = &clone
	return nil
}

func (r *fakeSessionRepo) GetByID(_ context.Context, id string) (*entity.Session, error) {
	s, ok := r.sessions[id]
	if !ok {
		return nil, stdErrors.New("session not found")
	}
	clone := *s
	return &clone, nil
}

func (r *fakeSessionRepo) GetByRefreshTokenHash(_ context.Context, hash string) (*entity.Session, error) {
	for _, s := range r.sessions {
		if s.RefreshTokenHash == hash {
			clone := *s
			return &clone, nil
		}
	}
	return nil, stdErrors.New("session not found")
}

func (r *fakeSessionRepo) UpdateRefreshToken(_ context.Context, id, newHash string, expiresAt time.Time, lastUsed time.Time) error {
	s, ok := r.sessions[id]
	if !ok {
		return stdErrors.New("session not found")
	}
	s.RefreshTokenHash = newHash
	s.RefreshTokenExpiresAt = expiresAt
	s.LastUsedAt = lastUsed
	s.RevokedAt = nil
	return nil
}

func (r *fakeSessionRepo) Revoke(_ context.Context, id string, revokedAt time.Time) error {
	s, ok := r.sessions[id]
	if !ok {
		return stdErrors.New("session not found")
	}
	s.RevokedAt = &revokedAt
	s.LastUsedAt = revokedAt
	return nil
}

func (r *fakeSessionRepo) RevokeByUser(_ context.Context, userID uint, revokedAt time.Time) error {
	for _, s := range r.sessions {
		if s.UserID == userID {
			s.RevokedAt = &revokedAt
			s.LastUsedAt = revokedAt
		}
	}
	return nil
}

func (r *fakeSessionRepo) ListActiveByUser(_ context.Context, userID uint, now time.Time) ([]*entity.Session, error) {
	var result []*entity.Session
	for _, s := range r.sessions {
		if s.UserID == userID && s.IsActive(now) {
			clone := *s
			result = append(result, &clone)
		}
	}
	return result, nil
}

func (r *fakeSessionRepo) CleanupExpired(_ context.Context, olderThan time.Time) error {
	for id, s := range r.sessions {
		if s.RefreshTokenExpiresAt.Before(olderThan) {
			delete(r.sessions, id)
		}
	}
	return nil
}

func (r *fakeUserRepo) Create(_ context.Context, user *entity.User) error {
	user.ID = r.nextID
	r.nextID++
	snapshot := *user
	r.byID[user.ID] = &snapshot
	r.byUsername[user.Username] = &snapshot
	r.byEmail[user.Email] = &snapshot
	return nil
}

func (r *fakeUserRepo) GetByID(_ context.Context, id uint) (*entity.User, error) {
	user, ok := r.byID[id]
	if !ok {
		return nil, errors.NewUserNotFoundError()
	}
	clone := *user
	return &clone, nil
}

func (r *fakeUserRepo) GetByUsername(_ context.Context, username string) (*entity.User, error) {
	user, ok := r.byUsername[username]
	if !ok {
		return nil, errors.NewUserNotFoundError()
	}
	clone := *user
	return &clone, nil
}

func (r *fakeUserRepo) GetByEmail(_ context.Context, email string) (*entity.User, error) {
	user, ok := r.byEmail[email]
	if !ok {
		return nil, errors.NewUserNotFoundError()
	}
	clone := *user
	return &clone, nil
}

func (r *fakeUserRepo) Update(_ context.Context, user *entity.User) error {
	_, ok := r.byID[user.ID]
	if !ok {
		return errors.NewUserNotFoundError()
	}
	snapshot := *user
	r.byID[user.ID] = &snapshot
	r.byUsername[user.Username] = &snapshot
	r.byEmail[user.Email] = &snapshot
	return nil
}

func (r *fakeUserRepo) Delete(_ context.Context, id uint) error { return nil }

func (r *fakeUserRepo) List(_ context.Context, _, _ int) ([]*entity.User, error) { return nil, nil }

func (r *fakeUserRepo) ExistsByUsername(_ context.Context, username string) (bool, error) {
	_, ok := r.byUsername[username]
	return ok, nil
}

func (r *fakeUserRepo) ExistsByEmail(_ context.Context, email string) (bool, error) {
	_, ok := r.byEmail[email]
	return ok, nil
}

func TestRegister(t *testing.T) {
	repo := newFakeUserRepo()
	sessionRepo := newFakeSessionRepo()
	tokenService := token.NewTokenService("secret", time.Minute, 2*time.Hour)
	svc := NewAuthService(repo, sessionRepo, tokenService)

	user, err := svc.Register(context.Background(), "johndoe", "john@example.com", "supersecret")
	require.NoError(t, err)
	require.Equal(t, "johndoe", user.Username)
	require.Equal(t, types.UserRoleUser, user.Role)

	_, err = svc.Register(context.Background(), "johndoe", "john+dup@example.com", "anotherpass")
	require.Error(t, err)
	domainErr, ok := err.(*errors.DomainError)
	require.True(t, ok)
	require.Equal(t, errors.CodeUserAlreadyExists, domainErr.Code)

	_, err = svc.Register(context.Background(), "janedoe", "john@example.com", "anotherpass")
	require.Error(t, err)
	domainErr, ok = err.(*errors.DomainError)
	require.True(t, ok)
	require.Equal(t, errors.CodeUserAlreadyExists, domainErr.Code)
}

func TestRegisterPasswordPolicy(t *testing.T) {
	repo := newFakeUserRepo()
	sessionRepo := newFakeSessionRepo()
	tokenService := token.NewTokenService("secret", time.Minute, 2*time.Hour)
	svc := NewAuthService(repo, sessionRepo, tokenService)

	_, err := svc.Register(context.Background(), "jd", "short@example.com", "short")
	require.Error(t, err)
	domainErr, ok := err.(*errors.DomainError)
	require.True(t, ok)
	require.Equal(t, errors.CodeInvalidPassword, domainErr.Code)
}

func TestLogin(t *testing.T) {
	repo := newFakeUserRepo()
	sessionRepo := newFakeSessionRepo()
	tokenService := token.NewTokenService("secret", time.Minute, 2*time.Hour)
	svc := NewAuthService(repo, sessionRepo, tokenService)

	_, err := svc.Register(context.Background(), "johndoe", "john@example.com", "supersecret")
	require.NoError(t, err)

	tokenPair, user, err := svc.Login(context.Background(), "johndoe", "supersecret", &LoginMetadata{IPAddress: "127.0.0.1", UserAgent: "test"})
	require.NoError(t, err)
	require.NotEmpty(t, tokenPair.AccessToken)
	require.NotEmpty(t, tokenPair.RefreshToken)
	require.NotEmpty(t, tokenPair.SessionID)
	require.True(t, tokenPair.RefreshExpiresIn > 0)
	require.Equal(t, "johndoe", user.Username)

	_, _, err = svc.Login(context.Background(), "johndoe", "wrongpass", &LoginMetadata{})
	require.Error(t, err)
	domainErr, ok := err.(*errors.DomainError)
	require.True(t, ok)
	require.Equal(t, errors.CodeInvalidCredentials, domainErr.Code)
}

func TestRefresh(t *testing.T) {
	repo := newFakeUserRepo()
	sessionRepo := newFakeSessionRepo()
	tokenService := token.NewTokenService("secret", time.Minute, time.Hour)
	svc := NewAuthService(repo, sessionRepo, tokenService)

	_, err := svc.Register(context.Background(), "johndoe", "john@example.com", "supersecret")
	require.NoError(t, err)

	loginPair, _, err := svc.Login(context.Background(), "johndoe", "supersecret", &LoginMetadata{})
	require.NoError(t, err)

	refreshed, _, err := svc.Refresh(context.Background(), loginPair.SessionID, loginPair.RefreshToken)
	require.NoError(t, err)
	require.NotEqual(t, loginPair.RefreshToken, refreshed.RefreshToken)
	require.Equal(t, loginPair.SessionID, refreshed.SessionID)

	_, _, err = svc.Refresh(context.Background(), loginPair.SessionID, loginPair.RefreshToken)
	require.Error(t, err)
	domainErr, ok := err.(*errors.DomainError)
	require.True(t, ok)
	require.Equal(t, errors.CodeTokenInvalid, domainErr.Code)
}

func TestLogout(t *testing.T) {
	repo := newFakeUserRepo()
	sessionRepo := newFakeSessionRepo()
	tokenService := token.NewTokenService("secret", time.Minute, time.Hour)
	svc := NewAuthService(repo, sessionRepo, tokenService)

	_, err := svc.Register(context.Background(), "johndoe", "john@example.com", "supersecret")
	require.NoError(t, err)

	loginPair, _, err := svc.Login(context.Background(), "johndoe", "supersecret", &LoginMetadata{})
	require.NoError(t, err)

	require.NoError(t, svc.Logout(context.Background(), loginPair.SessionID))

	session, err := sessionRepo.GetByID(context.Background(), loginPair.SessionID)
	require.NoError(t, err)
	require.NotNil(t, session.RevokedAt)
}

func TestLogoutAll(t *testing.T) {
	repo := newFakeUserRepo()
	sessionRepo := newFakeSessionRepo()
	tokenService := token.NewTokenService("secret", time.Minute, time.Hour)
	svc := NewAuthService(repo, sessionRepo, tokenService)

	_, err := svc.Register(context.Background(), "johndoe", "john@example.com", "supersecret")
	require.NoError(t, err)

	loginPair, user, err := svc.Login(context.Background(), "johndoe", "supersecret", &LoginMetadata{})
	require.NoError(t, err)

	require.NoError(t, svc.LogoutAll(context.Background(), user.ID))

	session, err := sessionRepo.GetByID(context.Background(), loginPair.SessionID)
	require.NoError(t, err)
	require.NotNil(t, session.RevokedAt)
}
