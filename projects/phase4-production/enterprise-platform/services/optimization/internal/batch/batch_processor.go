package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agent-learning/enterprise-platform/services/optimization/internal/model"
)

// BatchProcessor 批处理器
type BatchProcessor struct {
	batchSize    int
	flushTimeout time.Duration
	queue        []map[string]interface{}
	results      map[string]chan map[string]interface{}
	handler      BatchHandler
	mu           sync.Mutex
	timer        *time.Timer
}

// BatchHandler 批处理处理器
type BatchHandler func(context.Context, []map[string]interface{}) ([]map[string]interface{}, error)

// NewBatchProcessor 创建批处理器
func NewBatchProcessor(batchSize int, flushTimeout time.Duration, handler BatchHandler) *BatchProcessor {
	bp := &BatchProcessor{
		batchSize:    batchSize,
		flushTimeout: flushTimeout,
		queue:        make([]map[string]interface{}, 0, batchSize),
		results:      make(map[string]chan map[string]interface{}),
		handler:      handler,
	}

	return bp
}

// Add 添加请求到批处理队列
func (bp *BatchProcessor) Add(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	bp.mu.Lock()

	// 生成请求ID
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
	request["_batch_id"] = requestID

	// 创建结果通道
	resultChan := make(chan map[string]interface{}, 1)
	bp.results[requestID] = resultChan

	// 添加到队列
	bp.queue = append(bp.queue, request)

	// 检查是否需要立即刷新
	shouldFlush := len(bp.queue) >= bp.batchSize

	// 重置或启动定时器
	if bp.timer == nil {
		bp.timer = time.AfterFunc(bp.flushTimeout, func() {
			bp.flush(context.Background())
		})
	} else if shouldFlush {
		bp.timer.Stop()
	}

	bp.mu.Unlock()

	// 如果达到批大小，立即刷新
	if shouldFlush {
		bp.flush(ctx)
	}

	// 等待结果
	select {
	case result := <-resultChan:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for batch result")
	}
}

// flush 刷新批处理队列
func (bp *BatchProcessor) flush(ctx context.Context) {
	bp.mu.Lock()

	if len(bp.queue) == 0 {
		bp.mu.Unlock()
		return
	}

	// 获取当前队列
	batch := bp.queue
	bp.queue = make([]map[string]interface{}, 0, bp.batchSize)

	// 重置定时器
	if bp.timer != nil {
		bp.timer.Stop()
		bp.timer = nil
	}

	bp.mu.Unlock()

	// 执行批处理
	results, err := bp.handler(ctx, batch)

	bp.mu.Lock()
	defer bp.mu.Unlock()

	// 分发结果
	for i, request := range batch {
		requestID, ok := request["_batch_id"].(string)
		if !ok {
			continue
		}

		resultChan, ok := bp.results[requestID]
		if !ok {
			continue
		}

		var result map[string]interface{}
		if err != nil {
			result = map[string]interface{}{
				"error": err.Error(),
			}
		} else if i < len(results) {
			result = results[i]
		} else {
			result = map[string]interface{}{
				"error": "no result for request",
			}
		}

		select {
		case resultChan <- result:
		default:
		}

		close(resultChan)
		delete(bp.results, requestID)
	}
}

// RequestMerger 请求合并器
type RequestMerger struct {
	processors map[string]*BatchProcessor
	mu         sync.RWMutex
}

// NewRequestMerger 创建请求合并器
func NewRequestMerger() *RequestMerger {
	return &RequestMerger{
		processors: make(map[string]*BatchProcessor),
	}
}

// RegisterProcessor 注册批处理器
func (rm *RequestMerger) RegisterProcessor(key string, processor *BatchProcessor) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.processors[key] = processor
}

// Process 处理请求
func (rm *RequestMerger) Process(ctx context.Context, key string, request map[string]interface{}) (map[string]interface{}, error) {
	rm.mu.RLock()
	processor, ok := rm.processors[key]
	rm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no processor registered for key: %s", key)
	}

	return processor.Add(ctx, request)
}

// BatchManager 批处理管理器
type BatchManager struct {
	batches map[string]*model.BatchRequest
	mu      sync.RWMutex
}

// NewBatchManager 创建批处理管理器
func NewBatchManager() *BatchManager {
	return &BatchManager{
		batches: make(map[string]*model.BatchRequest),
	}
}

// CreateBatch 创建批次
func (bm *BatchManager) CreateBatch(batch *model.BatchRequest) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.batches[batch.ID]; exists {
		return fmt.Errorf("batch already exists: %s", batch.ID)
	}

	bm.batches[batch.ID] = batch
	return nil
}

// GetBatch 获取批次
func (bm *BatchManager) GetBatch(batchID string) (*model.BatchRequest, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	batch, ok := bm.batches[batchID]
	if !ok {
		return nil, fmt.Errorf("batch not found: %s", batchID)
	}

	return batch, nil
}

// UpdateBatchStatus 更新批次状态
func (bm *BatchManager) UpdateBatchStatus(batchID, status string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	batch, ok := bm.batches[batchID]
	if !ok {
		return fmt.Errorf("batch not found: %s", batchID)
	}

	batch.Status = status
	batch.UpdatedAt = time.Now()

	return nil
}

// AddBatchResults 添加批次结果
func (bm *BatchManager) AddBatchResults(batchID string, results []map[string]interface{}) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	batch, ok := bm.batches[batchID]
	if !ok {
		return fmt.Errorf("batch not found: %s", batchID)
	}

	batch.Results = results
	batch.Status = "completed"
	batch.UpdatedAt = time.Now()

	return nil
}

// DeleteBatch 删除批次
func (bm *BatchManager) DeleteBatch(batchID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, ok := bm.batches[batchID]; !ok {
		return fmt.Errorf("batch not found: %s", batchID)
	}

	delete(bm.batches, batchID)
	return nil
}

// ListBatches 列出批次
func (bm *BatchManager) ListBatches(tenantID string) []*model.BatchRequest {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	var batches []*model.BatchRequest
	for _, batch := range bm.batches {
		if batch.TenantID == tenantID {
			batches = append(batches, batch)
		}
	}

	return batches
}

// CleanupCompletedBatches 清理已完成的批次
func (bm *BatchManager) CleanupCompletedBatches(olderThan time.Duration) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)

	for id, batch := range bm.batches {
		if batch.Status == "completed" && batch.UpdatedAt.Before(cutoff) {
			delete(bm.batches, id)
		}
	}
}

// AutoCleanup 自动清理
func (bm *BatchManager) AutoCleanup(interval time.Duration, olderThan time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		bm.CleanupCompletedBatches(olderThan)
	}
}
