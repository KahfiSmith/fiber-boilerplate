package types

import "time"

type User struct {
	ID              uint      `json:"id" gorm:"primarykey"`
	Name            string    `json:"name" gorm:"type:varchar(120);not null" validate:"required,min=2"`
	Email           string    `json:"email" gorm:"type:varchar(255);uniqueIndex;not null" validate:"required,email"`
	PasswordHash    *string   `json:"-" gorm:"type:varchar(255)"`
	Role            string    `json:"role" gorm:"type:varchar(50);not null;default:'user'"`
	IsEmailVerified bool      `json:"is_email_verified" gorm:"default:false"`
	OAuthProvider   *string   `json:"-" gorm:"type:varchar(50);index:idx_oauth_provider_subject,unique"`
	OAuthSubject    *string   `json:"-" gorm:"type:varchar(255);index:idx_oauth_provider_subject,unique"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
