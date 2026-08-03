package response

import (
	"errors"
	"fmt"

	"fiber-boilerplate/src/common/exceptions"
	"github.com/gofiber/fiber/v3"
	"github.com/go-playground/validator/v10"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func Success(ctx fiber.Ctx, status int, message string, data interface{}) error {
	return ctx.Status(status).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func HandleError(ctx fiber.Ctx, err error) error {
	var statusCode = fiber.StatusInternalServerError
	var errorResponse interface{} = "Internal Server Error"

	// custom application exceptions
	var appErr *exceptions.HttpError
	if errors.As(err, &appErr) {
		statusCode = appErr.Code
		errorResponse = appErr.Message
	}

	// fiber errors (e.g., 404, 405)
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		statusCode = fiberErr.Code
		errorResponse = fiberErr.Message
	}

	// validator errors
	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		statusCode = fiber.StatusBadRequest
		errs := make(map[string]string)
		for _, e := range valErrs {
			errs[e.Field()] = fmt.Sprintf("failed on the '%s' tag", e.Tag())
		}
		errorResponse = errs
	}

	msg := "An error occurred"
	if strErr, ok := errorResponse.(string); ok {
		msg = strErr
	}

	return ctx.Status(statusCode).JSON(APIResponse{
		Success: false,
		Message: msg,
		Error:   errorResponse,
	})
}

// globalerrorhandler for fiber app
func GlobalErrorHandler(ctx fiber.Ctx, err error) error {
	return HandleError(ctx, err)
}
