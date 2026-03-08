package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"fiber-boilerplate/pkg/entities"
	serverMiddleware "fiber-boilerplate/pkg/server/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

func parseAndValidate(c fiber.Ctx, payload any) error {
	body := c.Body()
	if len(body) == 0 {
		return errors.New("request body is empty")
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(payload); err != nil {
		return fmt.Errorf("parse body: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}

	return validateRequestPayload(c, payload)
}

func validateRequestPayload(c fiber.Ctx, payload any) error {
	validateAny := c.Locals(serverMiddleware.ValidatorLocalKey)
	validate, ok := validateAny.(*validator.Validate)
	if !ok || validate == nil {
		return errors.New("validator is not available in request context")
	}

	if err := validate.Struct(payload); err != nil {
		return fmt.Errorf("validate request: %w", err)
	}

	return nil
}

func requestMeta(c fiber.Ctx) entities.AuthClientMeta {
	return entities.AuthClientMeta{
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
	}
}

func accessTokenFromRequest(c fiber.Ctx) (string, error) {
	return bearerToken(c.Get("Authorization"))
}

func bearerToken(authHeader string) (string, error) {
	trimmed := strings.TrimSpace(authHeader)
	if trimmed == "" {
		return "", errors.New("authorization header is empty")
	}

	parts := strings.SplitN(trimmed, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("authorization header must be Bearer token")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("bearer token is empty")
	}

	return token, nil
}
