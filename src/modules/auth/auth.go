package auth

import (
	"time"
	"fiber-boilerplate/src/common/response"
	"fiber-boilerplate/src/common/validator"
	"fiber-boilerplate/src/common/exceptions"
	"fiber-boilerplate/src/common/database"
	"github.com/gofiber/fiber/v3"
)

// --- Structs (Combined DTO, Entity, Model) ---
type User struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	Name         string    `json:"name" gorm:"type:varchar(255);not null" validate:"required,min=2"`
	Email        string    `json:"email" gorm:"type:varchar(255);uniqueIndex;not null" validate:"required,email"`
	PasswordHash string    `json:"-" gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AuthRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

// --- Controller ---
type AuthController struct {
	service *AuthService
}

func NewAuthController(service *AuthService) *AuthController {
	return &AuthController{service: service}
}

func (c *AuthController) Login(ctx fiber.Ctx) error {
	var req AuthRequest
	
	if err := validator.ParseAndValidate(ctx, &req); err != nil {
		return response.HandleError(ctx, err)
	}

	res, err := c.service.Login(req)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, fiber.StatusOK, res)
}

// --- Service ---
type AuthService struct {
	repo *AuthRepository
}

func NewAuthService(repo *AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Login(req AuthRequest) (AuthResponse, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return AuthResponse{}, exceptions.Unauthorized("Invalid credentials")
	}

	// Logic to compare password hash should go here
	// Logic to generate JWT should go here

	return AuthResponse{
		AccessToken:  "jwt_access_token",
		RefreshToken: "jwt_refresh_token",
		User:         *user,
	}, nil
}

// --- Repository ---
type AuthRepository struct{}

func NewAuthRepository() *AuthRepository {
	return &AuthRepository{}
}

func (r *AuthRepository) FindByEmail(email string) (*User, error) {
	var user User
	// Using global DB instance from common/database
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}


// --- Module Routes ---
func RegisterRoutes(router fiber.Router, controller *AuthController) {
	authGroup := router.Group("/auth")
	authGroup.Post("/login", controller.Login)
}
