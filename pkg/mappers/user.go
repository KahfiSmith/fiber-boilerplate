package mappers

import (
	"fiber-boilerplate/pkg/entities"
	"fiber-boilerplate/pkg/models"
)

func ToUserModel(user *entities.User) models.User {
	if user == nil {
		return models.User{}
	}

	return models.User{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func ToUserEntity(modelUser models.User) entities.User {
	return entities.User{
		ID:           modelUser.ID,
		Name:         modelUser.Name,
		Email:        modelUser.Email,
		PasswordHash: modelUser.PasswordHash,
		CreatedAt:    modelUser.CreatedAt,
		UpdatedAt:    modelUser.UpdatedAt,
	}
}

func ApplyUserModel(user *entities.User, modelUser models.User) {
	if user == nil {
		return
	}

	user.ID = modelUser.ID
	user.Name = modelUser.Name
	user.Email = modelUser.Email
	user.PasswordHash = modelUser.PasswordHash
	user.CreatedAt = modelUser.CreatedAt
	user.UpdatedAt = modelUser.UpdatedAt
}
