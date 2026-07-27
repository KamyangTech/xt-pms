package monitor

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ------ 请求级指标 ------

type requestCounter struct {
	total    atomic.Int64
	byStatus [6]atomic.Int64 // 1xx, 2xx, 3xx, 4xx, 5xx, unknown
}

func (c *requestCounter) record(status int) {
	c.total.Add(1)
	switch {
	case status >= 100 && status < 200:
		c.byStatus[0].Add(1)
	case status >= 200 && status < 300:
		c.byStatus[1].Add(1)
	case status >= 300 && status < 400:
		c.byStatus[2].Add(1)
	case status >= 400 && status < 500:
		c.byStatus[3].Add(1)
	case status >= 500 && status < 600:
		c.byStatus[4].Add(1)
	default:
		c.byStatus[5].Add(1)
	}
}

func (c *requestCounter) snapshot() (total int64, byStatus map[string]int64) {
	byStatus = map[string]int64{
		"1xx": c.byStatus[0].Load(),
		"2xx": c.byStatus[1].Load(),
		"3xx": c.byStatus[2].Load(),
		"4xx": c.byStatus[3].Load(),
		"5xx": c.byStatus[4].Load(),
	}
	total = c.total.Load()
	return
}

// ------ 延迟采样环形缓冲区 ------

const defaultSampleSize = 2048

type latencyRing struct {
	mu       sync.Mutex
	buf      []float64 // 毫秒
	pos      int
	filled   bool // 是否已满一轮
}

func newLatencyRing(size int) *latencyRing {
	if size <= 0 {
		size = defaultSampleSize
	}
	return &latencyRing{buf: make([]float64, size)}
}

func (r *latencyRing) record(ms float64) {
	r.mu.Lock()
	r.buf[r.pos] = ms
	r.pos++
	if r.pos >= len(r.buf) {
		r.pos = 0
		r.filled = true
	}
	r.mu.Unlock()
}

func (r *latencyRing) percentile(p float64) float64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	n := r.pos
	if r.filled {
		n = len(r.buf)
	}
	if n == 0 {
		return 0
	}

	clone := make([]float64, n)
	copy(clone, r.buf[:n])
	sort.Float64s(clone)

	idx := int(math.Ceil(p/100*float64(n))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	return clone[idx]
}

func (r *latencyRing) avg() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	n := r.pos
	if r.filled {
		n = len(r.buf)
	}
	if n == 0 {
		return 0
	}
	sum := 0.0
	for i := 0; i < n; i++ {
		sum += r.buf[i]
	}
	return sum / float64(n)
}

func (r *latencyRing) max() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	n := r.pos
	if r.filled {
		n = len(r.buf)
	}
	if n == 0 {
		return 0
	}
	m := r.buf[0]
	for i := 1; i < n; i++ {
		if r.buf[i] > m {
			m = r.buf[i]
		}
	}
	return m
}

func (r *latencyRing) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.filled {
		return len(r.buf)
	}
	return r.pos
}

// ------ 原子版 in-flight ------

type inFlightGauge struct {
	val atomic.Int64
}

func (g *inFlightGauge) add(delta int64) {
	g.val.Add(delta)
}

func (g *inFlightGauge) load() int64 {
	return g.val.Load()
}

// ------ RuntimeMetrics ------

type RuntimeMetrics struct {
	uptime    time.Time
	totalReq  requestCounter
	latency   *latencyRing
	slowQuery requestCounter // 超过 500ms 的请求
	inFlight  inFlightGauge
}

var globalMetrics = &RuntimeMetrics{
	uptime:  time.Now(),
	latency: newLatencyRing(defaultSampleSize),
}

// RecordRequest 记录一次 HTTP 请求
func RecordRequest(method, path string, status int, latency time.Duration) {
	globalMetrics.totalReq.record(status)
	globalMetrics.latency.record(float64(latency.Microseconds()) / 1000.0) // 存毫秒
	if latency > 500*time.Millisecond {
		globalMetrics.slowQuery.record(status)
	}
}

// IncInFlight / DecInFlight 增减正在处理的请求数
func IncInFlight()  { globalMetrics.inFlight.add(1) }
func DecInFlight()  { globalMetrics.inFlight.add(-1) }

// Snapshot 导出当前指标快照
func Snapshot() map[string]interface{} {
	total, byStatus := globalMetrics.totalReq.snapshot()
	slowTotal, _ := globalMetrics.slowQuery.snapshot()

	return map[string]interface{}{
		"uptime_seconds":     fmt.Sprintf("%.0f", time.Since(globalMetrics.uptime).Seconds()),
		"total_requests":     total,
		"requests_by_status": byStatus,
		"in_flight":          globalMetrics.inFlight.load(),
		"latency_ms": map[string]interface{}{
			"avg":   math.Round(globalMetrics.latency.avg()*100) / 100,
			"p50":   math.Round(globalMetrics.latency.percentile(50)*100) / 100,
			"p95":   math.Round(globalMetrics.latency.percentile(95)*100) / 100,
			"p99":   math.Round(globalMetrics.latency.percentile(99)*100) / 100,
			"max":   math.Round(globalMetrics.latency.max()*100) / 100,
			"count": globalMetrics.latency.count(),
		},
		"slow_requests_total": slowTotal,
	}
}
