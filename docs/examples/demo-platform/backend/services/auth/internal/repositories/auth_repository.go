package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/eggybyte-technology/go-eggybyte-core/db"
)

// Session represents the session data model.
type Session struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// TODO: Add your model fields here
}

// SessionRepository handles database operations for session models.
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new instance of SessionRepository.
func NewSessionRepository() *SessionRepository {
	return &SessionRepository{}
}

// TableName returns the database table name for this repository.
func (r *SessionRepository) TableName() string {
	return "sessions"
}

// InitTable performs table creation and schema migration.
func (r *SessionRepository) InitTable(ctx context.Context, database *gorm.DB) error {
	r.db = database
	return r.db.WithContext(ctx).AutoMigrate(&Session{})
}

// Create inserts a new Session record into the database.
func (r *SessionRepository) Create(ctx context.Context, model *Session) error {
	return r.db.WithContext(ctx).Create(model).Error
}

// FindByID retrieves a Session by its ID.
func (r *SessionRepository) FindByID(ctx context.Context, id uint) (*Session, error) {
	var model Session
	err := r.db.WithContext(ctx).First(&model, id).Error
	return &model, err
}

// Update modifies an existing Session record.
func (r *SessionRepository) Update(ctx context.Context, model *Session) error {
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a Session record by ID.
func (r *SessionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Session{}, id).Error
}

// init registers this repository for automatic table initialization.
func init() {
	db.RegisterRepository(NewSessionRepository())
}
