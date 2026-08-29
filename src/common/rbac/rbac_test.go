package rbac_test

import (
	"testing"

	"fiber-boilerplate/src/common/rbac"
)

func TestHasPermission(t *testing.T) {
	tests := []struct {
		role       string
		permission string
		expected   bool
	}{
		{"admin", "admin.access", true},
		{"admin", "user.delete", true},
		{"admin", "user.read", true},
		{"user", "user.read", true},
		{"user", "admin.access", false},
		{"user", "user.delete", false},
		{"guest", "user.read", false},
	}

	for _, tt := range tests {
		got := rbac.HasPermission(tt.role, tt.permission)
		if got != tt.expected {
			t.Errorf("HasPermission(%q, %q) = %v; want %v", tt.role, tt.permission, got, tt.expected)
		}
	}
}
