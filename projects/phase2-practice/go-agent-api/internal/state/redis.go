package state

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisStateManager manages agent and task state in Redis
type RedisStateManager struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisStateManager creates a new Redis state manager
func NewRedisStateManager(addr, password string, db int) (*RedisStateManager, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStateManager{
		client: client,
		ctx:    ctx,
	}, nil
}

// Close closes the Redis connection
func (r *RedisStateManager) Close() error {
	return r.client.Close()
}

// SetAgentState stores agent state in Redis
func (r *RedisStateManager) SetAgentState(agentID string, state interface{}) error {
	key := fmt.Sprintf("agent:%s:state", agentID)
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := r.client.Set(r.ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to set state: %w", err)
	}

	return nil
}

// GetAgentState retrieves agent state from Redis
func (r *RedisStateManager) GetAgentState(agentID string, state interface{}) error {
	key := fmt.Sprintf("agent:%s:state", agentID)
	data, err := r.client.Get(r.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("state not found for agent: %s", agentID)
		}
		return fmt.Errorf("failed to get state: %w", err)
	}

	if err := json.Unmarshal(data, state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return nil
}

// DeleteAgentState removes agent state from Redis
func (r *RedisStateManager) DeleteAgentState(agentID string) error {
	key := fmt.Sprintf("agent:%s:state", agentID)
	if err := r.client.Del(r.ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete state: %w", err)
	}
	return nil
}

// SetTaskState stores task state in Redis
func (r *RedisStateManager) SetTaskState(taskID string, state interface{}) error {
	key := fmt.Sprintf("task:%s:state", taskID)
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Set with expiration (24 hours)
	if err := r.client.Set(r.ctx, key, data, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to set state: %w", err)
	}

	return nil
}

// GetTaskState retrieves task state from Redis
func (r *RedisStateManager) GetTaskState(taskID string, state interface{}) error {
	key := fmt.Sprintf("task:%s:state", taskID)
	data, err := r.client.Get(r.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("state not found for task: %s", taskID)
		}
		return fmt.Errorf("failed to get state: %w", err)
	}

	if err := json.Unmarshal(data, state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return nil
}

// SetTaskStatus stores task status in Redis
func (r *RedisStateManager) SetTaskStatus(taskID, status string) error {
	key := fmt.Sprintf("task:%s:status", taskID)
	if err := r.client.Set(r.ctx, key, status, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to set status: %w", err)
	}
	return nil
}

// GetTaskStatus retrieves task status from Redis
func (r *RedisStateManager) GetTaskStatus(taskID string) (string, error) {
	key := fmt.Sprintf("task:%s:status", taskID)
	status, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("status not found for task: %s", taskID)
		}
		return "", fmt.Errorf("failed to get status: %w", err)
	}
	return status, nil
}

// ListAgents returns all agent IDs
func (r *RedisStateManager) ListAgents() ([]string, error) {
	keys, err := r.client.Keys(r.ctx, "agent:*:state").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	agents := make([]string, 0, len(keys))
	for _, key := range keys {
		// Extract agent ID from key format "agent:<id>:state"
		if len(key) > 13 {
			agentID := key[6 : len(key)-6]
			agents = append(agents, agentID)
		}
	}

	return agents, nil
}

// ListTasks returns all task IDs
func (r *RedisStateManager) ListTasks() ([]string, error) {
	keys, err := r.client.Keys(r.ctx, "task:*:state").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	tasks := make([]string, 0, len(keys))
	for _, key := range keys {
		// Extract task ID from key format "task:<id>:state"
		if len(key) > 12 {
			taskID := key[5 : len(key)-6]
			tasks = append(tasks, taskID)
		}
	}

	return tasks, nil
}
