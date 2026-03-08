package controllers

import (
	"fiber-boilerplate/pkg/dto/response"
	"fiber-boilerplate/pkg/entities"
)

const timeLayoutRFC3339 = "2006-01-02T15:04:05Z07:00"

func authTokenResponse(result entities.AuthTokens) response.AuthTokenResponse {
	return response.AuthTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresInSec: result.ExpiresInSec,
		SessionID:    result.SessionID,
		User:         authUserResponse(result.User),
	}
}

func otpChallengeResponse(result entities.OTPChallengeResult) response.OTPChallengeResponse {
	return response.OTPChallengeResponse{
		ChallengeID:  result.ChallengeID,
		ExpiresInSec: result.ExpiresInSec,
		OTP:          result.OTP,
	}
}

func authUserResponse(user entities.AuthUser) response.AuthUserResponse {
	return response.AuthUserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}

func authSessionResponses(sessions []entities.AuthSession) []response.AuthSessionResponse {
	resp := make([]response.AuthSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		resp = append(resp, response.AuthSessionResponse{
			SessionID: session.SessionID,
			UserAgent: session.UserAgent,
			IPAddress: session.IPAddress,
			CreatedAt: session.CreatedAt.Format(timeLayoutRFC3339),
			ExpiresAt: session.ExpiresAt.Format(timeLayoutRFC3339),
			Current:   session.Current,
		})
	}

	return resp
}
