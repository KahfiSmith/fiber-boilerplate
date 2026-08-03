package auth

import (
	"fiber-boilerplate/src/modules/auth/types"
	"fiber-boilerplate/src/database"
)

type AuthRepository struct{}

func NewAuthRepository() *AuthRepository {
	return &AuthRepository{}
}

func (r *AuthRepository) Create(user *types.User) error {
	return database.DB.Create(user).Error
}

func (r *AuthRepository) FindByEmail(email string) (*types.User, error) {
	var user types.User
	if err := database.DB.Where("LOWER(email) = LOWER(?)", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) UpdatePassword(userID uint, passwordHash string) error {
	return database.DB.Model(&types.User{}).Where("id = ?", userID).Update("password_hash", passwordHash).Error
}

func (r *AuthRepository) MarkEmailAsVerified(userID uint) error {
	return database.DB.Model(&types.User{}).Where("id = ?", userID).Update("is_email_verified", true).Error
}

func (r *AuthRepository) FindByID(id uint) (*types.User, error) {
	var user types.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) Delete(userID uint) error {
	return database.DB.Delete(&types.User{}, userID).Error
}