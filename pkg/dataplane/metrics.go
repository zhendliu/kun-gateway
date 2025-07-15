package dataplane

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics 监控指标
type Metrics struct {
	// 基础指标
	totalRequests  int64
	totalResponses int64
	activeRequests int64

	// 状态码统计
	statusCodes map[int]int64

	// 延迟统计
	latencySum   int64 // 纳秒
	latencyCount int64
	latencyMin   int64
	latencyMax   int64

	// 域名维度指标
	domainMetrics map[string]*DomainMetrics

	mu sync.RWMutex
}

// DomainMetrics 域名维度指标
type DomainMetrics struct {
	Requests     int64
	BytesIn      int64
	BytesOut     int64
	SuccessCount int64
	ErrorCount   int64
	LatencySum   int64
	LatencyCount int64
}

// NewMetrics 创建监控指标
func NewMetrics() *Metrics {
	return &Metrics{
		statusCodes:   make(map[int]int64),
		domainMetrics: make(map[string]*DomainMetrics),
	}
}

// IncRequests 增加请求数
func (m *Metrics) IncRequests() {
	atomic.AddInt64(&m.totalRequests, 1)
	atomic.AddInt64(&m.activeRequests, 1)
}

// IncResponses 增加响应数
func (m *Metrics) IncResponses() {
	atomic.AddInt64(&m.totalResponses, 1)
	atomic.AddInt64(&m.activeRequests, -1)
}

// IncStatusCodes 增加状态码计数
func (m *Metrics) IncStatusCodes(code int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusCodes[code]++
}

// RecordLatency 记录延迟
func (m *Metrics) RecordLatency(duration time.Duration) {
	ns := duration.Nanoseconds()

	atomic.AddInt64(&m.latencySum, ns)
	atomic.AddInt64(&m.latencyCount, 1)

	// 更新最小延迟
	for {
		old := atomic.LoadInt64(&m.latencyMin)
		if old != 0 && ns >= old {
			break
		}
		if atomic.CompareAndSwapInt64(&m.latencyMin, old, ns) {
			break
		}
	}

	// 更新最大延迟
	for {
		old := atomic.LoadInt64(&m.latencyMax)
		if ns <= old {
			break
		}
		if atomic.CompareAndSwapInt64(&m.latencyMax, old, ns) {
			break
		}
	}
}

// RecordDomainMetrics 记录域名维度指标
func (m *Metrics) RecordDomainMetrics(domain string, success bool, latency time.Duration, bytesIn, bytesOut int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	dm, exists := m.domainMetrics[domain]
	if !exists {
		dm = &DomainMetrics{}
		m.domainMetrics[domain] = dm
	}

	dm.Requests++
	dm.BytesIn += bytesIn
	dm.BytesOut += bytesOut
	dm.LatencySum += latency.Nanoseconds()
	dm.LatencyCount++

	if success {
		dm.SuccessCount++
	} else {
		dm.ErrorCount++
	}
}

// GetStats 获取统计信息
func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})

	// 基础指标
	stats["total_requests"] = atomic.LoadInt64(&m.totalRequests)
	stats["total_responses"] = atomic.LoadInt64(&m.totalResponses)
	stats["active_requests"] = atomic.LoadInt64(&m.activeRequests)

	// 延迟统计
	latencyCount := atomic.LoadInt64(&m.latencyCount)
	if latencyCount > 0 {
		latencySum := atomic.LoadInt64(&m.latencySum)
		latencyMin := atomic.LoadInt64(&m.latencyMin)
		latencyMax := atomic.LoadInt64(&m.latencyMax)

		stats["latency"] = map[string]interface{}{
			"avg_ms": float64(latencySum) / float64(latencyCount) / 1e6,
			"min_ms": float64(latencyMin) / 1e6,
			"max_ms": float64(latencyMax) / 1e6,
			"count":  latencyCount,
		}
	}

	// 状态码统计
	stats["status_codes"] = m.statusCodes

	// 域名维度统计
	stats["domains"] = m.domainMetrics

	return stats
}

// Reset 重置指标
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreInt64(&m.totalRequests, 0)
	atomic.StoreInt64(&m.totalResponses, 0)
	atomic.StoreInt64(&m.activeRequests, 0)
	atomic.StoreInt64(&m.latencySum, 0)
	atomic.StoreInt64(&m.latencyCount, 0)
	atomic.StoreInt64(&m.latencyMin, 0)
	atomic.StoreInt64(&m.latencyMax, 0)

	m.statusCodes = make(map[int]int64)
	m.domainMetrics = make(map[string]*DomainMetrics)
}
