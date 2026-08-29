package validator

import (
	"fmt"
	"strings"

	"fiber-boilerplate/src/common/exceptions"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

var validate = validator.New()

func msgForTag(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

func ParseAndValidate(c fiber.Ctx, payload any) error {
	if err := c.Bind().JSON(payload); err != nil {
		return exceptions.BadRequest("Invalid request body payload")
	}

	if err := validate.Struct(payload); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return exceptions.New(fiber.StatusBadRequest, "VALIDATION_ERROR", msgForTag(validationErrors[0]))
		}
		return exceptions.New(fiber.StatusBadRequest, "VALIDATION_ERROR", "Validation failed")
	}

	return nil
}
