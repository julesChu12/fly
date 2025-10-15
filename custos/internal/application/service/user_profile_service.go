package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/julesChu12/fly/custos/internal/application/dto"
	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
)

// UserProfileService handles user profile business logic
type UserProfileService struct {
	userRepo        repository.UserRepository
	userProfileRepo repository.UserProfileRepository
}

// NewUserProfileService creates a new user profile service
func NewUserProfileService(
	userRepo repository.UserRepository,
	userProfileRepo repository.UserProfileRepository,
) *UserProfileService {
	return &UserProfileService{
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
	}
}

// GetProfile retrieves a user's profile by user ID
func (s *UserProfileService) GetProfile(ctx context.Context, userID uint) (*dto.GetProfileResponse, error) {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get user profile, create if not exists
	profile, err := s.userProfileRepo.GetByUserID(ctx, userID)
	if err != nil {
		if err == repository.ErrUserProfileNotFound {
			// Auto-create profile with default values
			profile = entity.NewUserProfile(userID)
			if err := s.userProfileRepo.Create(ctx, profile); err != nil {
				return nil, fmt.Errorf("failed to create profile: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get profile: %w", err)
		}
	}

	// Build response
	response := &dto.GetProfileResponse{
		UserID:   profile.UserID,
		Nickname: profile.Nickname,
		Avatar:   profile.Avatar,
		Gender:   profile.Gender,
		Extra:    profile.Extra,
	}

	// Include user basic info
	response.Username = user.Username
	response.Email = user.Email

	if profile.Birthday != nil {
		response.Birthday = profile.Birthday.Format("2006-01-02")
	}

	return response, nil
}

// UpdateProfile updates a user's profile
func (s *UserProfileService) UpdateProfile(ctx context.Context, userID uint, req *dto.UpdateProfileRequest) error {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Get existing profile or create new one
	profile, err := s.userProfileRepo.GetByUserID(ctx, userID)
	profileExists := err == nil

	if err != nil && err != repository.ErrUserProfileNotFound {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if !profileExists {
		// Create new profile
		profile = entity.NewUserProfile(userID)
	}

	// Update fields
	if err := s.applyProfileUpdates(profile, req); err != nil {
		return err
	}

	// Save profile
	if !profileExists {
		if err := s.userProfileRepo.Create(ctx, profile); err != nil {
			return fmt.Errorf("failed to create profile: %w", err)
		}
	} else {
		if err := s.userProfileRepo.Update(ctx, profile); err != nil {
			return fmt.Errorf("failed to update profile: %w", err)
		}
	}

	return nil
}

// DeleteProfile deletes a user's profile
func (s *UserProfileService) DeleteProfile(ctx context.Context, userID uint) error {
	// Check if profile exists
	exists, err := s.userProfileRepo.Exists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check profile existence: %w", err)
	}

	if !exists {
		return repository.ErrUserProfileNotFound
	}

	// Delete profile
	if err := s.userProfileRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	return nil
}

// CreateProfile creates a new profile for a user
func (s *UserProfileService) CreateProfile(ctx context.Context, userID uint, req *dto.UpdateProfileRequest) error {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Check if profile already exists
	exists, err := s.userProfileRepo.Exists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check profile existence: %w", err)
	}

	if exists {
		return errors.New("profile already exists")
	}

	// Create new profile
	profile := entity.NewUserProfile(userID)
	if err := s.applyProfileUpdates(profile, req); err != nil {
		return err
	}

	if err := s.userProfileRepo.Create(ctx, profile); err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	return nil
}

// applyProfileUpdates applies update request to profile entity
func (s *UserProfileService) applyProfileUpdates(profile *entity.UserProfile, req *dto.UpdateProfileRequest) error {
	if req.Nickname != "" {
		profile.Nickname = req.Nickname
	}

	if req.Avatar != "" {
		profile.Avatar = req.Avatar
	}

	if req.Gender != "" {
		profile.Gender = req.Gender
	}

	if req.Birthday != "" {
		birthday, err := time.Parse("2006-01-02", req.Birthday)
		if err != nil {
			return fmt.Errorf("invalid birthday format, use YYYY-MM-DD: %w", err)
		}
		profile.Birthday = &birthday
	}

	if req.Extra != "" {
		profile.Extra = req.Extra
	}

	return nil
}
