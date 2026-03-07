package repository

import (
	"context"
	"errors"
	"fiber-boilerplate/pkg/entities"
	"fiber-boilerplate/pkg/mappers"
	"fiber-boilerplate/pkg/models"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrOTPChallengeNotFound = errors.New("otp challenge not found")
var ErrOTPInvalidCode = errors.New("otp code is invalid")
var ErrOTPTooManyAttempts = errors.New("otp too many attempts")

type OTPRepository interface {
	StoreChallenge(ctx context.Context, challenge entities.OTPChallenge, ttl time.Duration) error
	VerifyChallenge(ctx context.Context, challengeID string, codeHash string, purpose string) (uint, error)
}

type otpRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) OTPRepository {
	return &otpRepository{db: db}
}

func (r *otpRepository) StoreChallenge(ctx context.Context, challenge entities.OTPChallenge, ttl time.Duration) error {
	modelChallenge := mappers.ToOTPChallengeModel(challenge)
	if err := r.db.WithContext(ctx).Create(&modelChallenge).Error; err != nil {
		return fmt.Errorf("store otp challenge: %w", err)
	}

	return nil
}

func (r *otpRepository) VerifyChallenge(ctx context.Context, challengeID string, codeHash string, purpose string) (uint, error) {
	challengeID = strings.TrimSpace(challengeID)
	if challengeID == "" {
		return 0, ErrOTPChallengeNotFound
	}

	var userID uint
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var challenge models.OTPChallenge
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("challenge_id = ? AND purpose = ?", challengeID, strings.TrimSpace(purpose)).
			First(&challenge).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOTPChallengeNotFound
			}
			return fmt.Errorf("load otp challenge: %w", err)
		}

		now := time.Now().UTC()
		if !challenge.ExpiresAt.After(now) {
			if err := tx.Delete(&challenge).Error; err != nil {
				return fmt.Errorf("delete expired otp challenge: %w", err)
			}
			return ErrOTPChallengeNotFound
		}
		if challenge.Attempts >= challenge.MaxAttempts {
			if err := tx.Delete(&challenge).Error; err != nil {
				return fmt.Errorf("delete exhausted otp challenge: %w", err)
			}
			return ErrOTPTooManyAttempts
		}
		if challenge.CodeHash != codeHash {
			nextAttempts := challenge.Attempts + 1
			if nextAttempts >= challenge.MaxAttempts {
				if err := tx.Delete(&challenge).Error; err != nil {
					return fmt.Errorf("delete invalid otp challenge: %w", err)
				}
				return ErrOTPTooManyAttempts
			}

			if err := tx.Model(&challenge).Update("attempts", nextAttempts).Error; err != nil {
				return fmt.Errorf("persist otp attempt: %w", err)
			}
			return ErrOTPInvalidCode
		}

		userID = challenge.UserID
		if err := tx.Delete(&challenge).Error; err != nil {
			return fmt.Errorf("delete verified otp challenge: %w", err)
		}

		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrOTPChallengeNotFound):
			return 0, ErrOTPChallengeNotFound
		case errors.Is(err, ErrOTPInvalidCode):
			return 0, ErrOTPInvalidCode
		case errors.Is(err, ErrOTPTooManyAttempts):
			return 0, ErrOTPTooManyAttempts
		default:
			return 0, fmt.Errorf("verify otp challenge: %w", err)
		}
	}

	return userID, nil
}
