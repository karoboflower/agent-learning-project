package state

import (
	"sync"
	"time"
)

// StateManager manages application state
type StateManager struct {
	redis    *RedisStateManager
	cache    *MemoryCache
	enabled  bool
}

// MemoryCache provides in-memory caching
type MemoryCache struct {
	data map[string]CacheItem
	mu   sync.RWMutex
}

// CacheItem represents a cached item
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]CacheItem),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Set stores a value in cache
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Get retrieves a value from cache
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	return item.Value, true
}

// Delete removes a value from cache
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

// cleanup removes expired items
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.data {
			if now.After(item.ExpiresAt) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// NewStateManager creates a new state manager
func NewStateManager(redisAddr, redisPassword string, redisDB int) (*StateManager, error) {
	redis, err := NewRedisStateManager(redisAddr, redisPassword, redisDB)
	if err != nil {
		// If Redis connection fails, disable Redis and use only cache
		return &StateManager{
			redis:   nil,
			cache:   NewMemoryCache(),
			enabled: false,
		}, nil
	}

	return &StateManager{
		redis:   redis,
		cache:   NewMemoryCache(),
		enabled: true,
	}, nil
}

// SaveAgentState saves agent state
func (sm *StateManager) SaveAgentState(agentID string, state interface{}) error {
	// Cache first
	sm.cache.Set("agent:"+agentID, state, 5*time.Minute)

	// Persist to Redis if enabled
	if sm.enabled && sm.redis != nil {
		return sm.redis.SetAgentState(agentID, state)
	}

	return nil
}

// LoadAgentState loads agent state
func (sm *StateManager) LoadAgentState(agentID string, state interface{}) error {
	// Try cache first
	if cached, ok := sm.cache.Get("agent:" + agentID); ok {
		return copyInterface(cached, state)
	}

	// Try Redis if enabled
	if sm.enabled && sm.redis != nil {
		if err := sm.redis.GetAgentState(agentID, state); err == nil {
			// Update cache
			sm.cache.Set("agent:"+agentID, state, 5*time.Minute)
			return nil
		}
	}

	return nil
}

// SaveTaskState saves task state
func (sm *StateManager) SaveTaskState(taskID string, state interface{}) error {
	// Cache first
	sm.cache.Set("task:"+taskID, state, 10*time.Minute)

	// Persist to Redis if enabled
	if sm.enabled && sm.redis != nil {
		return sm.redis.SetTaskState(taskID, state)
	}

	return nil
}

// LoadTaskState loads task state
func (sm *StateManager) LoadTaskState(taskID string, state interface{}) error {
	// Try cache first
	if cached, ok := sm.cache.Get("task:" + taskID); ok {
		return copyInterface(cached, state)
	}

	// Try Redis if enabled
	if sm.enabled && sm.redis != nil {
		if err := sm.redis.GetTaskState(taskID, state); err == nil {
			// Update cache
			sm.cache.Set("task:"+taskID, state, 10*time.Minute)
			return nil
		}
	}

	return nil
}

// copyInterface copies interface values (simplified)
func copyInterface(src, dst interface{}) error {
	// In a real implementation, use reflection or JSON marshal/unmarshal
	// For now, just return nil
	return nil
}

// Close closes the state manager
func (sm *StateManager) Close() error {
	if sm.redis != nil {
		return sm.redis.Close()
	}
	return nil
}
