package response

type AuthUserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AuthTokenResponse struct {
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	TokenType    string           `json:"token_type"`
	ExpiresInSec int64            `json:"expires_in_sec"`
	SessionID    string           `json:"session_id"`
	User         AuthUserResponse `json:"user"`
}

type OTPChallengeResponse struct {
	ChallengeID  string `json:"challenge_id"`
	ExpiresInSec int64  `json:"expires_in_sec"`
	OTP          string `json:"otp,omitempty"`
}

type AuthSessionResponse struct {
	SessionID string `json:"session_id"`
	UserAgent string `json:"user_agent"`
	IPAddress string `json:"ip_address"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at"`
	Current   bool   `json:"current"`
}
