package validator

import (
	"fiber-boilerplate/src/common/exceptions"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

var validate = validator.New()

// ParseAndValidate parses the request body and validates the struct
func ParseAndValidate(c fiber.Ctx, payload any) error {
	if err := c.Bind().JSON(payload); err != nil {
		return exceptions.BadRequest("Invalid request body payload")
	}

	if err := validate.Struct(payload); err != nil {
		var errMsgs []string
		for _, err := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, fmt.Sprintf("Field '%s' failed on the '%s' tag", err.Field(), err.Tag()))
		}
		// We return the first validation error for simplicity
		return exceptions.BadRequest(errMsgs[0])
	}

	return nil
}
