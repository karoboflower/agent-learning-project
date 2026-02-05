package metrics

import (
	"runtime"
	"time"
)

// ResourceCollector 资源收集器
type ResourceCollector struct {
	collector *MetricsCollector
	interval  time.Duration
	stopChan  chan struct{}
}

// NewResourceCollector 创建资源收集器
func NewResourceCollector(collector *MetricsCollector, interval time.Duration) *ResourceCollector {
	return &ResourceCollector{
		collector: collector,
		interval:  interval,
		stopChan:  make(chan struct{}),
	}
}

// Start 开始收集
func (rc *ResourceCollector) Start() {
	ticker := time.NewTicker(rc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rc.collectMetrics()
		case <-rc.stopChan:
			return
		}
	}
}

// Stop 停止收集
func (rc *ResourceCollector) Stop() {
	close(rc.stopChan)
}

// collectMetrics 收集指标
func (rc *ResourceCollector) collectMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Goroutine数量
	rc.collector.GoroutinesActive.Set(float64(runtime.NumGoroutine()))

	// 堆内存
	rc.collector.HeapAllocBytes.Set(float64(m.HeapAlloc))
	rc.collector.HeapInuseBytes.Set(float64(m.HeapInuse))
	rc.collector.StackInuseBytes.Set(float64(m.StackInuse))

	// 内存使用
	rc.collector.MemoryUsage.WithLabelValues("service", "heap_alloc").Set(float64(m.HeapAlloc))
	rc.collector.MemoryUsage.WithLabelValues("service", "heap_sys").Set(float64(m.HeapSys))
	rc.collector.MemoryUsage.WithLabelValues("service", "stack_inuse").Set(float64(m.StackInuse))
	rc.collector.MemoryUsage.WithLabelValues("service", "total_alloc").Set(float64(m.TotalAlloc))
}
