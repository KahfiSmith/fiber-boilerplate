package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthRepositoryReturnsServiceNameAndUTCNow(t *testing.T) {
	t.Parallel()

	repo := NewHealthRepository("fiber-boilerplate")
	now := repo.NowUTC()

	assert.Equal(t, "fiber-boilerplate", repo.ServiceName())
	assert.Equal(t, "UTC", now.Location().String())
}
