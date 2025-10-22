package mysql

import (
	"context"
	"fmt"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	moradb "github.com/julesChu12/fly/mora/pkg/db"
)

type UserRepository struct {
	client *moradb.Client
}

// NewUserRepository creates UserRepository from mora db.Client (extracted from Database.Client())
func NewUserRepository(client *moradb.Client) repository.UserRepository {
	return &UserRepository{client: client}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	// Use mora's Create helper for simple operations
	return r.client.Create(ctx, user)
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	// Use mora's First helper
	err := r.client.First(ctx, &user, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.client.DB().WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.client.DB().WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	// Use mora's Save helper
	return r.client.Save(ctx, user)
}

func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	// Use mora's Delete helper
	return r.client.Delete(ctx, &entity.User{}, id)
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	var users []*entity.User
	// Use mora's Paginate helper
	err := r.client.Paginate(ctx, &users, (offset/limit)+1, limit)
	return users, err
}

func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	// Use mora's Exists helper
	return r.client.Exists(ctx, &entity.User{}, "username = ?", username)
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	// Use mora's Exists helper
	return r.client.Exists(ctx, &entity.User{}, "email = ?", email)
}

// Multi-tenant methods implementation

func (r *UserRepository) GetByIDWithTenant(ctx context.Context, id uint, tenantID uint) (*entity.User, error) {
	var user entity.User
	err := r.client.DB().WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsernameWithTenant(ctx context.Context, username string, tenantID uint) (*entity.User, error) {
	var user entity.User
	err := r.client.DB().WithContext(ctx).Where("username = ? AND tenant_id = ?", username, tenantID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmailWithTenant(ctx context.Context, email string, tenantID uint) (*entity.User, error) {
	var user entity.User
	err := r.client.DB().WithContext(ctx).Where("email = ? AND tenant_id = ?", email, tenantID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ListByTenant(ctx context.Context, tenantID uint, limit, offset int) ([]*entity.User, error) {
	var users []*entity.User
	err := r.client.DB().WithContext(ctx).Where("tenant_id = ?", tenantID).Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

func (r *UserRepository) ExistsByUsernameWithTenant(ctx context.Context, username string, tenantID uint) (bool, error) {
	return r.client.Exists(ctx, &entity.User{}, "username = ? AND tenant_id = ?", username, tenantID)
}

func (r *UserRepository) ExistsByEmailWithTenant(ctx context.Context, email string, tenantID uint) (bool, error) {
	return r.client.Exists(ctx, &entity.User{}, "email = ? AND tenant_id = ?", email, tenantID)
}

// ListWithFilter lists users with advanced filtering
func (r *UserRepository) ListWithFilter(ctx context.Context, filter *repository.UserListFilter, limit, offset int) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	query := r.client.DB().WithContext(ctx).Model(&entity.User{})

	// Apply filters
	if filter != nil {
		if filter.Status != nil && *filter.Status != "" {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.Role != nil && *filter.Role != "" {
			query = query.Where("role = ?", *filter.Role)
		}
		if filter.UserType != nil && *filter.UserType != "" {
			query = query.Where("user_type = ?", *filter.UserType)
		}
		if filter.TenantID != nil {
			query = query.Where("tenant_id = ?", *filter.TenantID)
		}
		if filter.Keyword != nil && *filter.Keyword != "" {
			keyword := "%" + *filter.Keyword + "%"
			query = query.Where("username LIKE ? OR email LIKE ?", keyword, keyword)
		}
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated results
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// CountByStatus counts users by status
func (r *UserRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.client.Count(ctx, &entity.User{}, &count, "status = ?", status)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by status: %w", err)
	}
	return count, nil
}

// CountByRole counts users grouped by role
func (r *UserRepository) CountByRole(ctx context.Context) (map[string]int64, error) {
	type Result struct {
		Role  string
		Count int64
	}
	var results []Result

	err := r.client.DB().WithContext(ctx).Model(&entity.User{}).
		Select("role, COUNT(*) as count").
		Group("role").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count users by role: %w", err)
	}

	counts := make(map[string]int64)
	for _, result := range results {
		counts[result.Role] = result.Count
	}
	return counts, nil
}

// CountByType counts users grouped by type
func (r *UserRepository) CountByType(ctx context.Context) (map[string]int64, error) {
	type Result struct {
		UserType string
		Count    int64
	}
	var results []Result

	err := r.client.DB().WithContext(ctx).Model(&entity.User{}).
		Select("user_type, COUNT(*) as count").
		Group("user_type").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count users by type: %w", err)
	}

	counts := make(map[string]int64)
	for _, result := range results {
		counts[result.UserType] = result.Count
	}
	return counts, nil
}

// CountNewUsers counts users created since a specific date
func (r *UserRepository) CountNewUsers(ctx context.Context, since string) (int64, error) {
	var count int64
	err := r.client.Count(ctx, &entity.User{}, &count, "created_at >= ?", since)
	if err != nil {
		return 0, fmt.Errorf("failed to count new users: %w", err)
	}
	return count, nil
}

// CountTotal counts total users
func (r *UserRepository) CountTotal(ctx context.Context) (int64, error) {
	var count int64
	err := r.client.Count(ctx, &entity.User{}, &count)
	if err != nil {
		return 0, fmt.Errorf("failed to count total users: %w", err)
	}
	return count, nil
}
