package observability

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Metrics struct {
	namespace string
	inflight  atomic.Int64
	series    sync.Map
}

type seriesKey struct {
	Method string
	Path   string
	Status int
}

type seriesValue struct {
	count        atomic.Uint64
	durationNanos atomic.Uint64
}

type seriesSnapshot struct {
	key              seriesKey
	count            uint64
	durationSeconds  float64
}

func NewMetrics(appName string) *Metrics {
	return &Metrics{
		namespace: metricNamespace(appName),
	}
}

func (m *Metrics) IncInflight() {
	m.inflight.Add(1)
}

func (m *Metrics) DecInflight() {
	m.inflight.Add(-1)
}

func (m *Metrics) Observe(method string, path string, status int, duration time.Duration) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = "UNKNOWN"
	}

	path = strings.TrimSpace(path)
	if path == "" {
		path = "/"
	}

	key := seriesKey{
		Method: method,
		Path:   path,
		Status: status,
	}
	valueAny, _ := m.series.LoadOrStore(key, &seriesValue{})
	value := valueAny.(*seriesValue)
	value.count.Add(1)
	value.durationNanos.Add(uint64(duration))
}

func (m *Metrics) Handle(c fiber.Ctx) error {
	c.Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	c.Set("Cache-Control", "no-store")
	return c.SendString(m.Prometheus())
}

func (m *Metrics) Prometheus() string {
	var builder strings.Builder

	inflightMetric := m.metricName("http_requests_inflight")
	requestsMetric := m.metricName("http_requests_total")
	durationMetric := m.metricName("http_request_duration_seconds_total")
	goroutinesMetric := m.metricName("go_goroutines")
	allocMetric := m.metricName("go_memstats_alloc_bytes")
	heapInUseMetric := m.metricName("go_memstats_heap_inuse_bytes")
	gcMetric := m.metricName("go_gc_cycles_total")

	writeMetricHeader(&builder, inflightMetric, "Current in-flight HTTP requests.", "gauge")
	builder.WriteString(fmt.Sprintf("%s %d\n", inflightMetric, m.inflight.Load()))

	writeMetricHeader(&builder, requestsMetric, "Total HTTP requests processed.", "counter")
	writeMetricHeader(&builder, durationMetric, "Total accumulated HTTP request duration in seconds.", "counter")

	for _, snapshot := range m.snapshotSeries() {
		labels := fmt.Sprintf(`method="%s",path="%s",status="%d"`,
			escapeLabelValue(snapshot.key.Method),
			escapeLabelValue(snapshot.key.Path),
			snapshot.key.Status,
		)
		builder.WriteString(fmt.Sprintf("%s{%s} %d\n", requestsMetric, labels, snapshot.count))
		builder.WriteString(fmt.Sprintf("%s{%s} %.9f\n", durationMetric, labels, snapshot.durationSeconds))
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	writeMetricHeader(&builder, goroutinesMetric, "Current number of goroutines.", "gauge")
	builder.WriteString(fmt.Sprintf("%s %d\n", goroutinesMetric, runtime.NumGoroutine()))

	writeMetricHeader(&builder, allocMetric, "Current allocated memory in bytes.", "gauge")
	builder.WriteString(fmt.Sprintf("%s %d\n", allocMetric, mem.Alloc))

	writeMetricHeader(&builder, heapInUseMetric, "Current heap memory in use in bytes.", "gauge")
	builder.WriteString(fmt.Sprintf("%s %d\n", heapInUseMetric, mem.HeapInuse))

	writeMetricHeader(&builder, gcMetric, "Total completed GC cycles.", "counter")
	builder.WriteString(fmt.Sprintf("%s %d\n", gcMetric, mem.NumGC))

	return builder.String()
}

func (m *Metrics) metricName(suffix string) string {
	return m.namespace + "_" + suffix
}

func (m *Metrics) snapshotSeries() []seriesSnapshot {
	snapshots := make([]seriesSnapshot, 0)
	m.series.Range(func(key any, value any) bool {
		seriesKeyValue, ok := key.(seriesKey)
		if !ok {
			return true
		}

		seriesValueValue, ok := value.(*seriesValue)
		if !ok {
			return true
		}

		snapshots = append(snapshots, seriesSnapshot{
			key:             seriesKeyValue,
			count:           seriesValueValue.count.Load(),
			durationSeconds: float64(seriesValueValue.durationNanos.Load()) / float64(time.Second),
		})

		return true
	})

	sort.Slice(snapshots, func(i int, j int) bool {
		if snapshots[i].key.Path != snapshots[j].key.Path {
			return snapshots[i].key.Path < snapshots[j].key.Path
		}
		if snapshots[i].key.Method != snapshots[j].key.Method {
			return snapshots[i].key.Method < snapshots[j].key.Method
		}
		return snapshots[i].key.Status < snapshots[j].key.Status
	})

	return snapshots
}

func writeMetricHeader(builder *strings.Builder, name string, help string, metricType string) {
	builder.WriteString(fmt.Sprintf("# HELP %s %s\n", name, help))
	builder.WriteString(fmt.Sprintf("# TYPE %s %s\n", name, metricType))
}

func escapeLabelValue(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`)
	return replacer.Replace(value)
}

func metricNamespace(appName string) string {
	var builder strings.Builder
	lastUnderscore := false

	for _, r := range strings.ToLower(strings.TrimSpace(appName)) {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastUnderscore = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastUnderscore = false
		default:
			if !lastUnderscore && builder.Len() > 0 {
				builder.WriteByte('_')
				lastUnderscore = true
			}
		}
	}

	namespace := strings.Trim(builder.String(), "_")
	if namespace == "" {
		return "app"
	}
	if namespace[0] >= '0' && namespace[0] <= '9' {
		return "app_" + namespace
	}

	return namespace
}
