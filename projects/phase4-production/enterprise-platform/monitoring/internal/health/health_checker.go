package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck 健康检查接口
type HealthCheck interface {
	Name() string
	Check(ctx context.Context) error
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Name      string       `json:"name"`
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
}

// HealthReport 健康报告
type HealthReport struct {
	Status      HealthStatus                  `json:"status"`
	Timestamp   time.Time                     `json:"timestamp"`
	Duration    time.Duration                 `json:"duration"`
	Checks      map[string]*HealthCheckResult `json:"checks"`
	Version     string                        `json:"version"`
	BuildTime   string                        `json:"build_time"`
	Uptime      time.Duration                 `json:"uptime"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	checks    map[string]HealthCheck
	mu        sync.RWMutex
	startTime time.Time
	version   string
	buildTime string
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(version, buildTime string) *HealthChecker {
	return &HealthChecker{
		checks:    make(map[string]HealthCheck),
		startTime: time.Now(),
		version:   version,
		buildTime: buildTime,
	}
}

// Register 注册健康检查
func (hc *HealthChecker) Register(check HealthCheck) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.checks[check.Name()] = check
}

// Unregister 注销健康检查
func (hc *HealthChecker) Unregister(name string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	delete(hc.checks, name)
}

// Check 执行健康检查
func (hc *HealthChecker) Check(ctx context.Context) *HealthReport {
	start := time.Now()

	hc.mu.RLock()
	checks := make(map[string]HealthCheck, len(hc.checks))
	for name, check := range hc.checks {
		checks[name] = check
	}
	hc.mu.RUnlock()

	results := make(map[string]*HealthCheckResult)
	var wg sync.WaitGroup

	for name, check := range checks {
		wg.Add(1)

		go func(name string, check HealthCheck) {
			defer wg.Done()

			result := hc.runCheck(ctx, check)
			hc.mu.Lock()
			results[name] = result
			hc.mu.Unlock()
		}(name, check)
	}

	wg.Wait()

	// 确定整体状态
	overallStatus := StatusHealthy
	for _, result := range results {
		if result.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
			break
		} else if result.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	return &HealthReport{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		Checks:    results,
		Version:   hc.version,
		BuildTime: hc.buildTime,
		Uptime:    time.Since(hc.startTime),
	}
}

// runCheck 运行单个检查
func (hc *HealthChecker) runCheck(ctx context.Context, check HealthCheck) *HealthCheckResult {
	start := time.Now()

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 执行检查
	err := check.Check(ctx)

	result := &HealthCheckResult{
		Name:      check.Name(),
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}

	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = err.Error()
	} else {
		result.Status = StatusHealthy
		result.Message = "OK"
	}

	return result
}

// DatabaseHealthCheck 数据库健康检查
type DatabaseHealthCheck struct {
	name string
	ping func(context.Context) error
}

// NewDatabaseHealthCheck 创建数据库健康检查
func NewDatabaseHealthCheck(name string, ping func(context.Context) error) *DatabaseHealthCheck {
	return &DatabaseHealthCheck{
		name: name,
		ping: ping,
	}
}

// Name 返回检查名称
func (dhc *DatabaseHealthCheck) Name() string {
	return dhc.name
}

// Check 执行检查
func (dhc *DatabaseHealthCheck) Check(ctx context.Context) error {
	return dhc.ping(ctx)
}

// RedisHealthCheck Redis健康检查
type RedisHealthCheck struct {
	name string
	ping func(context.Context) error
}

// NewRedisHealthCheck 创建Redis健康检查
func NewRedisHealthCheck(name string, ping func(context.Context) error) *RedisHealthCheck {
	return &RedisHealthCheck{
		name: name,
		ping: ping,
	}
}

// Name 返回检查名称
func (rhc *RedisHealthCheck) Name() string {
	return rhc.name
}

// Check 执行检查
func (rhc *RedisHealthCheck) Check(ctx context.Context) error {
	return rhc.ping(ctx)
}

// MessageQueueHealthCheck 消息队列健康检查
type MessageQueueHealthCheck struct {
	name   string
	status func(context.Context) error
}

// NewMessageQueueHealthCheck 创建消息队列健康检查
func NewMessageQueueHealthCheck(name string, status func(context.Context) error) *MessageQueueHealthCheck {
	return &MessageQueueHealthCheck{
		name:   name,
		status: status,
	}
}

// Name 返回检查名称
func (mqhc *MessageQueueHealthCheck) Name() string {
	return mqhc.name
}

// Check 执行检查
func (mqhc *MessageQueueHealthCheck) Check(ctx context.Context) error {
	return mqhc.status(ctx)
}

// DiskSpaceHealthCheck 磁盘空间健康检查
type DiskSpaceHealthCheck struct {
	path      string
	threshold float64 // 阈值（百分比）
}

// NewDiskSpaceHealthCheck 创建磁盘空间健康检查
func NewDiskSpaceHealthCheck(path string, threshold float64) *DiskSpaceHealthCheck {
	return &DiskSpaceHealthCheck{
		path:      path,
		threshold: threshold,
	}
}

// Name 返回检查名称
func (dshc *DiskSpaceHealthCheck) Name() string {
	return fmt.Sprintf("disk_space_%s", dshc.path)
}

// Check 执行检查
func (dshc *DiskSpaceHealthCheck) Check(ctx context.Context) error {
	// 这里简化实现，实际应该使用syscall获取磁盘使用情况
	// 示例：如果磁盘使用超过阈值，返回错误
	// usage := getDiskUsage(dshc.path)
	// if usage > dshc.threshold {
	//     return fmt.Errorf("disk usage %.2f%% exceeds threshold %.2f%%", usage, dshc.threshold)
	// }
	return nil
}

// MemoryHealthCheck 内存健康检查
type MemoryHealthCheck struct {
	threshold uint64 // 阈值（字节）
}

// NewMemoryHealthCheck 创建内存健康检查
func NewMemoryHealthCheck(threshold uint64) *MemoryHealthCheck {
	return &MemoryHealthCheck{
		threshold: threshold,
	}
}

// Name 返回检查名称
func (mhc *MemoryHealthCheck) Name() string {
	return "memory"
}

// Check 执行检查
func (mhc *MemoryHealthCheck) Check(ctx context.Context) error {
	// 这里可以检查内存使用情况
	// 示例：如果内存使用超过阈值，返回错误
	return nil
}

// HTTPEndpointHealthCheck HTTP端点健康检查
type HTTPEndpointHealthCheck struct {
	name     string
	url      string
	checkFn  func(context.Context, string) error
}

// NewHTTPEndpointHealthCheck 创建HTTP端点健康检查
func NewHTTPEndpointHealthCheck(name, url string, checkFn func(context.Context, string) error) *HTTPEndpointHealthCheck {
	return &HTTPEndpointHealthCheck{
		name:    name,
		url:     url,
		checkFn: checkFn,
	}
}

// Name 返回检查名称
func (hehc *HTTPEndpointHealthCheck) Name() string {
	return hehc.name
}

// Check 执行检查
func (hehc *HTTPEndpointHealthCheck) Check(ctx context.Context) error {
	return hehc.checkFn(ctx, hehc.url)
}

// CompositeHealthCheck 组合健康检查
type CompositeHealthCheck struct {
	name   string
	checks []HealthCheck
}

// NewCompositeHealthCheck 创建组合健康检查
func NewCompositeHealthCheck(name string, checks ...HealthCheck) *CompositeHealthCheck {
	return &CompositeHealthCheck{
		name:   name,
		checks: checks,
	}
}

// Name 返回检查名称
func (chc *CompositeHealthCheck) Name() string {
	return chc.name
}

// Check 执行检查
func (chc *CompositeHealthCheck) Check(ctx context.Context) error {
	for _, check := range chc.checks {
		if err := check.Check(ctx); err != nil {
			return fmt.Errorf("%s: %w", check.Name(), err)
		}
	}
	return nil
}

// LivenessProbe 存活探针
type LivenessProbe struct {
	healthy bool
	mu      sync.RWMutex
}

// NewLivenessProbe 创建存活探针
func NewLivenessProbe() *LivenessProbe {
	return &LivenessProbe{
		healthy: true,
	}
}

// Name 返回探针名称
func (lp *LivenessProbe) Name() string {
	return "liveness"
}

// Check 执行检查
func (lp *LivenessProbe) Check(ctx context.Context) error {
	lp.mu.RLock()
	defer lp.mu.RUnlock()

	if !lp.healthy {
		return fmt.Errorf("service is not alive")
	}
	return nil
}

// SetHealthy 设置健康状态
func (lp *LivenessProbe) SetHealthy(healthy bool) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	lp.healthy = healthy
}

// ReadinessProbe 就绪探针
type ReadinessProbe struct {
	ready bool
	mu    sync.RWMutex
}

// NewReadinessProbe 创建就绪探针
func NewReadinessProbe() *ReadinessProbe {
	return &ReadinessProbe{
		ready: false,
	}
}

// Name 返回探针名称
func (rp *ReadinessProbe) Name() string {
	return "readiness"
}

// Check 执行检查
func (rp *ReadinessProbe) Check(ctx context.Context) error {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	if !rp.ready {
		return fmt.Errorf("service is not ready")
	}
	return nil
}

// SetReady 设置就绪状态
func (rp *ReadinessProbe) SetReady(ready bool) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	rp.ready = ready
}

// StartupProbe 启动探针
type StartupProbe struct {
	started bool
	mu      sync.RWMutex
}

// NewStartupProbe 创建启动探针
func NewStartupProbe() *StartupProbe {
	return &StartupProbe{
		started: false,
	}
}

// Name 返回探针名称
func (sp *StartupProbe) Name() string {
	return "startup"
}

// Check 执行检查
func (sp *StartupProbe) Check(ctx context.Context) error {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	if !sp.started {
		return fmt.Errorf("service has not started")
	}
	return nil
}

// SetStarted 设置启动状态
func (sp *StartupProbe) SetStarted(started bool) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.started = started
}
