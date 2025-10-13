package mysql

import (
	"context"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase(dsn string, debug bool) (*Database, error) {
	config := &gorm.Config{}
	if debug {
		config.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) AutoMigrate() error {
	return d.db.AutoMigrate(
		&entity.Tenant{},  // Add Tenant table
		&entity.User{},
		&entity.Session{},
	)
}

func (d *Database) DB() *gorm.DB {
	return d.db
}

func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, id).Error
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	var users []*entity.User
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// Multi-tenant methods implementation

func (r *UserRepository) GetByIDWithTenant(ctx context.Context, id uint, tenantID uint) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsernameWithTenant(ctx context.Context, username string, tenantID uint) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("username = ? AND tenant_id = ?", username, tenantID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmailWithTenant(ctx context.Context, email string, tenantID uint) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ? AND tenant_id = ?", email, tenantID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ListByTenant(ctx context.Context, tenantID uint, limit, offset int) ([]*entity.User, error) {
	var users []*entity.User
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

func (r *UserRepository) ExistsByUsernameWithTenant(ctx context.Context, username string, tenantID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("username = ? AND tenant_id = ?", username, tenantID).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) ExistsByEmailWithTenant(ctx context.Context, email string, tenantID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("email = ? AND tenant_id = ?", email, tenantID).Count(&count).Error
	return count > 0, err
}

// ListWithFilter lists users with advanced filtering
func (r *UserRepository) ListWithFilter(ctx context.Context, filter *repository.UserListFilter, limit, offset int) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.User{})

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
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("status = ?", status).Count(&count).Error
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

	err := r.db.WithContext(ctx).Model(&entity.User{}).
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

	err := r.db.WithContext(ctx).Model(&entity.User{}).
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
	err := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("created_at >= ?", since).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count new users: %w", err)
	}
	return count, nil
}

// CountTotal counts total users
func (r *UserRepository) CountTotal(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count total users: %w", err)
	}
	return count, nil
}
