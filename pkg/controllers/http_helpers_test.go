package controllers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"fiber-boilerplate/pkg/dto/request"
	serverMiddleware "fiber-boilerplate/pkg/server/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBearerToken(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		header  string
		want    string
		wantErr string
	}{
		{
			name:   "valid bearer token",
			header: "Bearer token-123",
			want:   "token-123",
		},
		{
			name:    "missing header",
			header:  "",
			wantErr: "authorization header is empty",
		},
		{
			name:    "wrong auth scheme",
			header:  "Basic token-123",
			wantErr: "authorization header must be Bearer token",
		},
		{
			name:    "empty token",
			header:  "Bearer   ",
			wantErr: "bearer token is empty",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			token, err := bearerToken(testCase.header)
			if testCase.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, testCase.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.want, token)
		})
	}
}

func TestParseAndValidate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		body            string
		withValidator   bool
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "valid payload",
			body:            `{"name":"Kahfi","email":"kahfi@example.com","password":"Secret123"}`,
			withValidator:   true,
			expectedStatus:  fiber.StatusNoContent,
			expectedMessage: "",
		},
		{
			name:            "unknown field rejected",
			body:            `{"name":"Kahfi","email":"kahfi@example.com","password":"Secret123","extra":"x"}`,
			withValidator:   true,
			expectedStatus:  fiber.StatusBadRequest,
			expectedMessage: "parse body",
		},
		{
			name:            "missing validator",
			body:            `{"name":"Kahfi","email":"kahfi@example.com","password":"Secret123"}`,
			withValidator:   false,
			expectedStatus:  fiber.StatusInternalServerError,
			expectedMessage: "validator is not available",
		},
		{
			name:            "invalid payload",
			body:            `{"name":"K","email":"not-an-email","password":"123"}`,
			withValidator:   true,
			expectedStatus:  fiber.StatusBadRequest,
			expectedMessage: "validate request",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			if testCase.withValidator {
				app.Use(serverMiddleware.InjectRequestContext(validator.New()))
			}

			app.Post("/", func(c fiber.Ctx) error {
				var payload request.RegisterRequest
				err := parseAndValidate(c, &payload)
				if err != nil {
					if strings.Contains(err.Error(), "validator is not available") {
						return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
					}
					return c.Status(fiber.StatusBadRequest).SendString(err.Error())
				}
				return c.SendStatus(fiber.StatusNoContent)
			})

			req := httptest.NewRequest("POST", "/", strings.NewReader(testCase.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedStatus, resp.StatusCode)

			if testCase.expectedMessage == "" {
				return
			}

			body := readBody(t, resp)
			assert.Contains(t, body, testCase.expectedMessage)
		})
	}
}
