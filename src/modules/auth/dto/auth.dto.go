package dto

import "fiber-boilerplate/src/modules/auth/types"

type AuthRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	ResetToken  string `json:"reset_token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type DeleteAccountRequest struct {
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	AccessToken string     `json:"access_token"`
	ExpiresIn   int        `json:"expires_in"`
	User        types.User `json:"user"`
}

type SessionResponse struct {
	SessionID  string    `json:"session_id"`
	IssuedAt   string    `json:"issued_at"`
	LastUsedAt string    `json:"last_used_at"`
	ExpiresAt  string    `json:"expires_at"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Current    bool      `json:"current"`
}
