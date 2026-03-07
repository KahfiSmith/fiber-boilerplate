package request

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=120"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type VerifyOTPRequest struct {
	ChallengeID string `json:"challenge_id" validate:"required,min=8,max=256"`
	OTP         string `json:"otp" validate:"required,len=6,numeric"`
}

type ResetPasswordRequest struct {
	ChallengeID string `json:"challenge_id" validate:"required,min=8,max=256"`
	OTP         string `json:"otp" validate:"required,len=6,numeric"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,min=32,max=512"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,min=32,max=512"`
}

type RevokeSessionRequest struct {
	SessionID string `json:"session_id" validate:"required,min=8,max=256"`
}
