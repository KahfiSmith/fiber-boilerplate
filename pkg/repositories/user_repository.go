package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"fiber-boilerplate/pkg/entities"
	"fiber-boilerplate/pkg/mappers"
	"fiber-boilerplate/pkg/models"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("user already exists")

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByID(ctx context.Context, id uint) (*entities.User, error)
	UpdatePassword(ctx context.Context, id uint, passwordHash string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	modelUser := mappers.ToUserModel(user)

	err := r.db.WithContext(ctx).Create(&modelUser).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}

	mappers.ApplyUserModel(user, modelUser)

	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	var modelUser models.User

	err := r.db.WithContext(ctx).
		Where("LOWER(email) = ?", strings.ToLower(strings.TrimSpace(email))).
		First(&modelUser).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	user := mappers.ToUserEntity(modelUser)
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*entities.User, error) {
	var modelUser models.User

	err := r.db.WithContext(ctx).First(&modelUser, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	user := mappers.ToUserEntity(modelUser)
	return &user, nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, id uint, passwordHash string) error {
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("password_hash", passwordHash)
	if result.Error != nil {
		return fmt.Errorf("update user password: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
