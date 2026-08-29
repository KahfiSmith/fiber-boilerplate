package rbac

import (
	"fiber-boilerplate/src/common/exceptions"
	"fiber-boilerplate/src/common/response"

	"github.com/gofiber/fiber/v3"
)

var RolePermissions = map[string][]string{
	"admin": {
		"user.read",
		"user.create",
		"user.update",
		"user.delete",
		"admin.access",
	},
	"user": {
		"user.read",
		"user.update",
	},
}

func HasPermission(role string, permission string) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

func RequirePermission(permission string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok || userRole == "" {
			return response.HandleError(c, exceptions.Forbidden("FORBIDDEN", "Forbidden: role required"))
		}

		if !HasPermission(userRole, permission) {
			return response.HandleError(c, exceptions.Forbidden("FORBIDDEN", "Forbidden: insufficient permissions"))
		}

		return c.Next()
	}
}
