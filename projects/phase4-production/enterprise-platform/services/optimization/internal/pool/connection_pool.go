package pool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/agent-learning/enterprise-platform/services/optimization/internal/model"
)

var (
	ErrPoolClosed    = errors.New("connection pool is closed")
	ErrPoolExhausted = errors.New("connection pool is exhausted")
	ErrInvalidConn   = errors.New("invalid connection")
)

// Connection 连接接口
type Connection interface {
	Close() error
	IsValid() bool
	Reset() error
}

// ConnectionFactory 连接工厂
type ConnectionFactory func() (Connection, error)

// ConnectionPool 连接池
type ConnectionPool struct {
	factory     ConnectionFactory
	minSize     int
	maxSize     int
	maxLifetime time.Duration
	maxIdleTime time.Duration
	conns       chan *PooledConnection
	mu          sync.RWMutex
	closed      bool
	stats       *model.ConnectionPoolStats
	config      *model.ConnectionPool
}

// PooledConnection 池化连接
type PooledConnection struct {
	conn       Connection
	pool       *ConnectionPool
	createdAt  time.Time
	lastUsedAt time.Time
	usageCount int64
}

// NewConnectionPool 创建连接池
func NewConnectionPool(config *model.ConnectionPool, factory ConnectionFactory) (*ConnectionPool, error) {
	if config.MinSize < 0 || config.MaxSize <= 0 || config.MinSize > config.MaxSize {
		return nil, fmt.Errorf("invalid pool size: min=%d, max=%d", config.MinSize, config.MaxSize)
	}

	pool := &ConnectionPool{
		factory:     factory,
		minSize:     config.MinSize,
		maxSize:     config.MaxSize,
		maxLifetime: config.MaxLifetime,
		maxIdleTime: config.MaxIdleTime,
		conns:       make(chan *PooledConnection, config.MaxSize),
		closed:      false,
		stats: &model.ConnectionPoolStats{
			TotalConns:    0,
			AcquiredConns: 0,
			ReleasedConns: 0,
			AvgWaitTime:   0,
			MaxWaitTime:   0,
		},
		config: config,
	}

	// 创建最小连接数
	for i := 0; i < config.MinSize; i++ {
		conn, err := pool.createConnection()
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("failed to create initial connection: %w", err)
		}
		pool.conns <- conn
	}

	// 启动清理goroutine
	go pool.cleanup()

	return pool, nil
}

// Acquire 获取连接
func (p *ConnectionPool) Acquire(ctx context.Context) (*PooledConnection, error) {
	if p.isClosed() {
		return nil, ErrPoolClosed
	}

	start := time.Now()

	select {
	case conn := <-p.conns:
		// 从池中获取连接
		if conn.isExpired() || !conn.conn.IsValid() {
			conn.Close()
			return p.Acquire(ctx) // 递归获取新连接
		}

		conn.lastUsedAt = time.Now()
		conn.usageCount++

		p.recordAcquire(time.Since(start))

		return conn, nil

	case <-ctx.Done():
		return nil, ctx.Err()

	default:
		// 池中无连接，尝试创建新连接
		p.mu.Lock()
		currentSize := p.getCurrentSize()

		if currentSize < p.maxSize {
			p.mu.Unlock()

			conn, err := p.createConnection()
			if err != nil {
				return nil, err
			}

			p.recordAcquire(time.Since(start))

			return conn, nil
		}
		p.mu.Unlock()

		// 达到最大连接数，等待可用连接
		select {
		case conn := <-p.conns:
			if conn.isExpired() || !conn.conn.IsValid() {
				conn.Close()
				return p.Acquire(ctx)
			}

			conn.lastUsedAt = time.Now()
			conn.usageCount++

			p.recordAcquire(time.Since(start))

			return conn, nil

		case <-ctx.Done():
			return nil, ctx.Err()

		case <-time.After(10 * time.Second):
			return nil, ErrPoolExhausted
		}
	}
}

// Release 释放连接
func (p *ConnectionPool) Release(conn *PooledConnection) error {
	if p.isClosed() {
		conn.Close()
		return ErrPoolClosed
	}

	if conn == nil || conn.conn == nil {
		return ErrInvalidConn
	}

	// 重置连接
	if err := conn.conn.Reset(); err != nil {
		conn.Close()
		return err
	}

	// 检查连接是否过期或失效
	if conn.isExpired() || !conn.conn.IsValid() {
		conn.Close()
		return nil
	}

	conn.lastUsedAt = time.Now()

	p.recordRelease()

	// 放回池中
	select {
	case p.conns <- conn:
		return nil
	default:
		// 池满了，关闭连接
		conn.Close()
		return nil
	}
}

// Close 关闭连接池
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	// 关闭所有连接
	close(p.conns)
	for conn := range p.conns {
		conn.Close()
	}

	return nil
}

// GetStats 获取统计信息
func (p *ConnectionPool) GetStats() *model.ConnectionPoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := *p.stats
	return &stats
}

// GetConfig 获取配置
func (p *ConnectionPool) GetConfig() *model.ConnectionPool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	config := *p.config
	config.CurrentSize = p.getCurrentSize()
	config.ActiveConns = p.maxSize - len(p.conns)
	config.IdleConns = len(p.conns)
	config.Stats = *p.stats

	return &config
}

// createConnection 创建连接
func (p *ConnectionPool) createConnection() (*PooledConnection, error) {
	conn, err := p.factory()
	if err != nil {
		return nil, err
	}

	pooledConn := &PooledConnection{
		conn:       conn,
		pool:       p,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
		usageCount: 0,
	}

	p.mu.Lock()
	p.stats.TotalConns++
	p.mu.Unlock()

	return pooledConn, nil
}

// getCurrentSize 获取当前连接数
func (p *ConnectionPool) getCurrentSize() int {
	return int(p.stats.TotalConns - (p.stats.ReleasedConns - p.stats.AcquiredConns))
}

// isClosed 检查是否已关闭
func (p *ConnectionPool) isClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.closed
}

// cleanup 清理过期连接
func (p *ConnectionPool) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if p.isClosed() {
			return
		}

		// 从池中取出连接检查
		var validConns []*PooledConnection

		for {
			select {
			case conn := <-p.conns:
				if conn.isExpired() || !conn.conn.IsValid() {
					conn.Close()
				} else {
					validConns = append(validConns, conn)
				}
			default:
				// 放回有效连接
				for _, conn := range validConns {
					select {
					case p.conns <- conn:
					default:
						conn.Close()
					}
				}
				goto nextIteration
			}
		}

	nextIteration:
	}
}

// recordAcquire 记录获取
func (p *ConnectionPool) recordAcquire(waitTime time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stats.AcquiredConns++

	// 更新平均等待时间
	if p.stats.AvgWaitTime == 0 {
		p.stats.AvgWaitTime = waitTime
	} else {
		p.stats.AvgWaitTime = (p.stats.AvgWaitTime*9 + waitTime) / 10
	}

	// 更新最大等待时间
	if waitTime > p.stats.MaxWaitTime {
		p.stats.MaxWaitTime = waitTime
	}

	p.config.WaitCount++
	p.config.WaitTime += waitTime
}

// recordRelease 记录释放
func (p *ConnectionPool) recordRelease() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stats.ReleasedConns++
}

// isExpired 检查连接是否过期
func (pc *PooledConnection) isExpired() bool {
	now := time.Now()

	// 检查最大生命周期
	if pc.pool.maxLifetime > 0 && now.Sub(pc.createdAt) > pc.pool.maxLifetime {
		return true
	}

	// 检查最大空闲时间
	if pc.pool.maxIdleTime > 0 && now.Sub(pc.lastUsedAt) > pc.pool.maxIdleTime {
		return true
	}

	return false
}

// Close 关闭池化连接
func (pc *PooledConnection) Close() error {
	if pc.conn != nil {
		return pc.conn.Close()
	}
	return nil
}

// PoolManager 连接池管理器
type PoolManager struct {
	pools map[string]*ConnectionPool
	mu    sync.RWMutex
}

// NewPoolManager 创建连接池管理器
func NewPoolManager() *PoolManager {
	return &PoolManager{
		pools: make(map[string]*ConnectionPool),
	}
}

// RegisterPool 注册连接池
func (pm *PoolManager) RegisterPool(name string, pool *ConnectionPool) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.pools[name]; exists {
		return fmt.Errorf("pool already registered: %s", name)
	}

	pm.pools[name] = pool
	return nil
}

// GetPool 获取连接池
func (pm *PoolManager) GetPool(name string) (*ConnectionPool, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pool, ok := pm.pools[name]
	if !ok {
		return nil, fmt.Errorf("pool not found: %s", name)
	}

	return pool, nil
}

// ClosePool 关闭连接池
func (pm *PoolManager) ClosePool(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pool, ok := pm.pools[name]
	if !ok {
		return fmt.Errorf("pool not found: %s", name)
	}

	delete(pm.pools, name)
	return pool.Close()
}

// CloseAll 关闭所有连接池
func (pm *PoolManager) CloseAll() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var errs []error
	for name, pool := range pm.pools {
		if err := pool.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close pool %s: %w", name, err))
		}
	}

	pm.pools = make(map[string]*ConnectionPool)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing pools: %v", errs)
	}

	return nil
}

// GetAllStats 获取所有连接池统计
func (pm *PoolManager) GetAllStats() map[string]*model.ConnectionPool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := make(map[string]*model.ConnectionPool)
	for name, pool := range pm.pools {
		stats[name] = pool.GetConfig()
	}

	return stats
}
