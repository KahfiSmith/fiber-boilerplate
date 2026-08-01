package observability

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetricNamespace(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "fiber_boilerplate", metricNamespace("Fiber Boilerplate"))
	assert.Equal(t, "app_123_api", metricNamespace("123 API"))
	assert.Equal(t, "app", metricNamespace("   "))
}

func TestPrometheusOutputIncludesSeriesAndRuntimeMetrics(t *testing.T) {
	t.Parallel()

	metrics := NewMetrics("fiber-boilerplate")
	metrics.IncInflight()
	metrics.Observe("GET", "/api/v1/health", 200, 25*time.Millisecond)
	metrics.Observe("POST", "/api/v1/auth/login", 401, 50*time.Millisecond)

	output := metrics.Prometheus()

	assert.True(t, strings.Contains(output, "fiber_boilerplate_http_requests_inflight 1"))
	assert.True(t, strings.Contains(output, `fiber_boilerplate_http_requests_total{method="GET",path="/api/v1/health",status="200"} 1`))
	assert.True(t, strings.Contains(output, `fiber_boilerplate_http_requests_total{method="POST",path="/api/v1/auth/login",status="401"} 1`))
	assert.True(t, strings.Contains(output, "fiber_boilerplate_http_request_duration_seconds_total"))
	assert.True(t, strings.Contains(output, "fiber_boilerplate_go_goroutines"))
}
