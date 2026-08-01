package auth

import (
	"time"
	"fiber-boilerplate/src/common/response"
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
	// if err := ctx.Bind().JSON(&req); err != nil {
	// 	return response.Error(ctx, fiber.StatusBadRequest, "Invalid request", err.Error())
	// }

	res, err := c.service.Login(req)
	if err != nil {
		return response.Error(ctx, fiber.StatusUnauthorized, "Login failed", err.Error())
	}

	return response.Success(ctx, fiber.StatusOK, res)
}

// --- Service ---
type AuthService struct {
	// repo *AuthRepository
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Login(req AuthRequest) (AuthResponse, error) {
	// Dummy logic
	return AuthResponse{
		AccessToken:  "dummy_access",
		RefreshToken: "dummy_refresh",
		User: User{
			ID:    1,
			Name:  "Test",
			Email: req.Email,
		},
	}, nil
}

// --- Module Routes ---
func RegisterRoutes(router fiber.Router, controller *AuthController) {
	authGroup := router.Group("/auth")
	authGroup.Post("/login", controller.Login)
}
