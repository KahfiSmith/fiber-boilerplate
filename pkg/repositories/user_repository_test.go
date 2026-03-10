package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeStoredEmail(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "kahfi@example.com", normalizeStoredEmail("  Kahfi@Example.com  "))
	assert.Equal(t, "", normalizeStoredEmail("   "))
}
