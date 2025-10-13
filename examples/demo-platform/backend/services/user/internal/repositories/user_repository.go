package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/eggybyte-technology/go-eggybyte-core/db"
)

// User represents the user data model.
type User struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// TODO: Add your model fields here
}

// UserRepository handles database operations for user models.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// TableName returns the database table name for this repository.
func (r *UserRepository) TableName() string {
	return "users"
}

// InitTable performs table creation and schema migration.
func (r *UserRepository) InitTable(ctx context.Context, database *gorm.DB) error {
	r.db = database
	return r.db.WithContext(ctx).AutoMigrate(&User{})
}

// Create inserts a new User record into the database.
func (r *UserRepository) Create(ctx context.Context, model *User) error {
	return r.db.WithContext(ctx).Create(model).Error
}

// FindByID retrieves a User by its ID.
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
	var model User
	err := r.db.WithContext(ctx).First(&model, id).Error
	return &model, err
}

// Update modifies an existing User record.
func (r *UserRepository) Update(ctx context.Context, model *User) error {
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a User record by ID.
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}

// init registers this repository for automatic table initialization.
func init() {
	db.RegisterRepository(NewUserRepository())
}
