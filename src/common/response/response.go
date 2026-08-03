package response

import (
	"errors"
	"fmt"

	"fiber-boilerplate/src/common/exceptions"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Code    string      `json:"code,omitempty"`
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
	var codeStr string
	var errorResponse interface{} = "Internal Server Error"

	var appErr *exceptions.HttpError
	if errors.As(err, &appErr) {
		statusCode = appErr.Code
		codeStr = appErr.CodeStr
		errorResponse = appErr.Message
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		statusCode = fiberErr.Code
		errorResponse = fiberErr.Message
	}

	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		statusCode = fiber.StatusBadRequest
		codeStr = "VALIDATION_ERROR"
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
		Code:    codeStr,
		Message: msg,
		Error:   errorResponse,
	})
}

func GlobalErrorHandler(ctx fiber.Ctx, err error) error {
	return HandleError(ctx, err)
}
