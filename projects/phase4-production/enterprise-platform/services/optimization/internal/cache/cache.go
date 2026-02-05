package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/agent-learning/enterprise-platform/services/optimization/internal/model"
)

// Cache 缓存接口
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, bool, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	GetStats() *model.CacheStats
}

// InMemoryCache 内存缓存
type InMemoryCache struct {
	entries map[string]*model.CacheEntry
	mu      sync.RWMutex
	stats   *model.CacheStats
	statsMu sync.RWMutex
}

// NewInMemoryCache 创建内存缓存
func NewInMemoryCache() *InMemoryCache {
	cache := &InMemoryCache{
		entries: make(map[string]*model.CacheEntry),
		stats: &model.CacheStats{
			TotalKeys:     0,
			TotalHits:     0,
			TotalMisses:   0,
			HitRate:       0,
			AvgAccessTime: 0,
			MemoryUsage:   0,
			EvictionCount: 0,
		},
	}

	// 启动清理goroutine
	go cache.cleanupExpired()

	return cache
}

// Get 获取缓存
func (c *InMemoryCache) Get(ctx context.Context, key string) (interface{}, bool, error) {
	start := time.Now()
	defer func() {
		c.updateAccessTime(time.Since(start))
	}()

	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		c.recordMiss()
		return nil, false, nil
	}

	// 检查是否过期
	if time.Now().After(entry.ExpiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()

		c.recordMiss()
		c.decrementKeyCount()
		return nil, false, nil
	}

	// 更新访问时间和命中次数
	c.mu.Lock()
	entry.Hits++
	entry.AccessedAt = time.Now()
	c.mu.Unlock()

	c.recordHit()

	return entry.Value, true, nil
}

// Set 设置缓存
func (c *InMemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查key是否已存在
	_, exists := c.entries[key]

	now := time.Now()
	entry := &model.CacheEntry{
		Key:        key,
		Value:      value.(map[string]interface{}),
		TTL:        ttl,
		Hits:       0,
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
		AccessedAt: now,
	}

	c.entries[key] = entry

	if !exists {
		c.incrementKeyCount()
	}

	return nil
}

// Delete 删除缓存
func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.entries[key]; exists {
		delete(c.entries, key)
		c.decrementKeyCount()
	}

	return nil
}

// Clear 清空缓存
func (c *InMemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*model.CacheEntry)

	c.statsMu.Lock()
	c.stats.TotalKeys = 0
	c.statsMu.Unlock()

	return nil
}

// GetStats 获取统计信息
func (c *InMemoryCache) GetStats() *model.CacheStats {
	c.statsMu.RLock()
	defer c.statsMu.RUnlock()

	// 复制统计信息
	stats := *c.stats

	// 计算命中率
	total := stats.TotalHits + stats.TotalMisses
	if total > 0 {
		stats.HitRate = float64(stats.TotalHits) / float64(total) * 100
	}

	return &stats
}

// cleanupExpired 清理过期条目
func (c *InMemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()

		now := time.Now()
		for key, entry := range c.entries {
			if now.After(entry.ExpiresAt) {
				delete(c.entries, key)
				c.incrementEvictionCount()
				c.decrementKeyCount()
			}
		}

		c.mu.Unlock()
	}
}

// recordHit 记录命中
func (c *InMemoryCache) recordHit() {
	c.statsMu.Lock()
	c.stats.TotalHits++
	c.statsMu.Unlock()
}

// recordMiss 记录未命中
func (c *InMemoryCache) recordMiss() {
	c.statsMu.Lock()
	c.stats.TotalMisses++
	c.statsMu.Unlock()
}

// updateAccessTime 更新访问时间
func (c *InMemoryCache) updateAccessTime(duration time.Duration) {
	c.statsMu.Lock()
	defer c.statsMu.Unlock()

	// 计算滑动平均
	if c.stats.AvgAccessTime == 0 {
		c.stats.AvgAccessTime = float64(duration.Microseconds()) / 1000.0
	} else {
		c.stats.AvgAccessTime = (c.stats.AvgAccessTime*0.9 + float64(duration.Microseconds())/1000.0*0.1)
	}
}

// incrementKeyCount 增加key计数
func (c *InMemoryCache) incrementKeyCount() {
	c.statsMu.Lock()
	c.stats.TotalKeys++
	c.statsMu.Unlock()
}

// decrementKeyCount 减少key计数
func (c *InMemoryCache) decrementKeyCount() {
	c.statsMu.Lock()
	if c.stats.TotalKeys > 0 {
		c.stats.TotalKeys--
	}
	c.statsMu.Unlock()
}

// incrementEvictionCount 增加驱逐计数
func (c *InMemoryCache) incrementEvictionCount() {
	c.statsMu.Lock()
	c.stats.EvictionCount++
	c.statsMu.Unlock()
}

// CacheManager 缓存管理器
type CacheManager struct {
	cache Cache
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(cache Cache) *CacheManager {
	return &CacheManager{
		cache: cache,
	}
}

// GetOrCompute 获取或计算
func (cm *CacheManager) GetOrCompute(ctx context.Context, key string, ttl time.Duration, compute func() (interface{}, error)) (interface{}, error) {
	// 尝试从缓存获取
	value, found, err := cm.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if found {
		return value, nil
	}

	// 计算值
	computed, err := compute()
	if err != nil {
		return nil, err
	}

	// 存入缓存
	if err := cm.cache.Set(ctx, key, computed, ttl); err != nil {
		// 记录错误但不影响返回
		fmt.Printf("failed to cache result: %v\n", err)
	}

	return computed, nil
}

// GenerateKey 生成缓存key
func (cm *CacheManager) GenerateKey(parts ...interface{}) string {
	data, _ := json.Marshal(parts)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// InvalidatePattern 按模式失效缓存
func (cm *CacheManager) InvalidatePattern(ctx context.Context, pattern string) error {
	// 这里简化实现，实际应该使用更高效的方法
	return cm.cache.Clear(ctx)
}

// Warmup 预热缓存
func (cm *CacheManager) Warmup(ctx context.Context, entries map[string]interface{}, ttl time.Duration) error {
	for key, value := range entries {
		if err := cm.cache.Set(ctx, key, value, ttl); err != nil {
			return fmt.Errorf("failed to warmup cache for key %s: %w", key, err)
		}
	}
	return nil
}

// LRUCache LRU缓存
type LRUCache struct {
	maxSize  int
	entries  map[string]*lruEntry
	head     *lruEntry
	tail     *lruEntry
	mu       sync.RWMutex
	stats    *model.CacheStats
	statsMu  sync.RWMutex
}

type lruEntry struct {
	key    string
	value  *model.CacheEntry
	prev   *lruEntry
	next   *lruEntry
}

// NewLRUCache 创建LRU缓存
func NewLRUCache(maxSize int) *LRUCache {
	cache := &LRUCache{
		maxSize: maxSize,
		entries: make(map[string]*lruEntry),
		stats: &model.CacheStats{
			TotalKeys:     0,
			TotalHits:     0,
			TotalMisses:   0,
			HitRate:       0,
			AvgAccessTime: 0,
			MemoryUsage:   0,
			EvictionCount: 0,
		},
	}

	// 创建虚拟头尾节点
	cache.head = &lruEntry{}
	cache.tail = &lruEntry{}
	cache.head.next = cache.tail
	cache.tail.prev = cache.head

	return cache
}

// Get 获取缓存
func (c *LRUCache) Get(ctx context.Context, key string) (interface{}, bool, error) {
	start := time.Now()
	defer func() {
		c.updateAccessTime(time.Since(start))
	}()

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if !exists {
		c.recordMiss()
		return nil, false, nil
	}

	// 检查是否过期
	if time.Now().After(entry.value.ExpiresAt) {
		c.removeEntry(entry)
		c.recordMiss()
		return nil, false, nil
	}

	// 移到链表头部
	c.moveToHead(entry)

	// 更新访问信息
	entry.value.Hits++
	entry.value.AccessedAt = time.Now()

	c.recordHit()

	return entry.value.Value, true, nil
}

// Set 设置缓存
func (c *LRUCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否已存在
	if entry, exists := c.entries[key]; exists {
		// 更新值
		entry.value.Value = value.(map[string]interface{})
		entry.value.ExpiresAt = time.Now().Add(ttl)
		entry.value.TTL = ttl

		// 移到头部
		c.moveToHead(entry)

		return nil
	}

	// 创建新条目
	now := time.Now()
	cacheEntry := &model.CacheEntry{
		Key:        key,
		Value:      value.(map[string]interface{}),
		TTL:        ttl,
		Hits:       0,
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
		AccessedAt: now,
	}

	entry := &lruEntry{
		key:   key,
		value: cacheEntry,
	}

	c.entries[key] = entry
	c.addToHead(entry)

	c.statsMu.Lock()
	c.stats.TotalKeys++
	c.statsMu.Unlock()

	// 检查是否超过容量
	if len(c.entries) > c.maxSize {
		c.removeTail()
	}

	return nil
}

// Delete 删除缓存
func (c *LRUCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.entries[key]; exists {
		c.removeEntry(entry)
	}

	return nil
}

// Clear 清空缓存
func (c *LRUCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*lruEntry)
	c.head.next = c.tail
	c.tail.prev = c.head

	c.statsMu.Lock()
	c.stats.TotalKeys = 0
	c.statsMu.Unlock()

	return nil
}

// GetStats 获取统计信息
func (c *LRUCache) GetStats() *model.CacheStats {
	c.statsMu.RLock()
	defer c.statsMu.RUnlock()

	stats := *c.stats

	total := stats.TotalHits + stats.TotalMisses
	if total > 0 {
		stats.HitRate = float64(stats.TotalHits) / float64(total) * 100
	}

	return &stats
}

// addToHead 添加到链表头部
func (c *LRUCache) addToHead(entry *lruEntry) {
	entry.next = c.head.next
	entry.prev = c.head
	c.head.next.prev = entry
	c.head.next = entry
}

// removeEntry 移除条目
func (c *LRUCache) removeEntry(entry *lruEntry) {
	entry.prev.next = entry.next
	entry.next.prev = entry.prev
	delete(c.entries, entry.key)

	c.statsMu.Lock()
	if c.stats.TotalKeys > 0 {
		c.stats.TotalKeys--
	}
	c.statsMu.Unlock()
}

// moveToHead 移到头部
func (c *LRUCache) moveToHead(entry *lruEntry) {
	entry.prev.next = entry.next
	entry.next.prev = entry.prev
	c.addToHead(entry)
}

// removeTail 移除尾部
func (c *LRUCache) removeTail() {
	tail := c.tail.prev
	if tail != c.head {
		c.removeEntry(tail)

		c.statsMu.Lock()
		c.stats.EvictionCount++
		c.statsMu.Unlock()
	}
}

// Helper methods
func (c *LRUCache) recordHit() {
	c.statsMu.Lock()
	c.stats.TotalHits++
	c.statsMu.Unlock()
}

func (c *LRUCache) recordMiss() {
	c.statsMu.Lock()
	c.stats.TotalMisses++
	c.statsMu.Unlock()
}

func (c *LRUCache) updateAccessTime(duration time.Duration) {
	c.statsMu.Lock()
	defer c.statsMu.Unlock()

	if c.stats.AvgAccessTime == 0 {
		c.stats.AvgAccessTime = float64(duration.Microseconds()) / 1000.0
	} else {
		c.stats.AvgAccessTime = (c.stats.AvgAccessTime*0.9 + float64(duration.Microseconds())/1000.0*0.1)
	}
}
