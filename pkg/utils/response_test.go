package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuccess(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return Success(c, fiber.StatusCreated, map[string]any{"name": "Kahfi"})
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	body := readResponseBody(t, resp)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	assert.Contains(t, body, `"success":true`)
	assert.Contains(t, body, `"name":"Kahfi"`)
}

func TestError(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return Error(c, fiber.StatusBadRequest, "invalid request", "missing field")
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	body := readResponseBody(t, resp)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, body, `"success":false`)
	assert.Contains(t, body, `"message":"invalid request"`)
	assert.Contains(t, body, `"error":"missing field"`)
}
