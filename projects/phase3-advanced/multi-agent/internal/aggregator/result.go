package aggregator

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ResultStatus 结果状态
type ResultStatus string

const (
	ResultStatusPending   ResultStatus = "PENDING"   // 待处理
	ResultStatusValidated ResultStatus = "VALIDATED" // 已验证
	ResultStatusRejected  ResultStatus = "REJECTED"  // 已拒绝
	ResultStatusMerged    ResultStatus = "MERGED"    // 已合并
)

// TaskResult Agent任务执行结果
type TaskResult struct {
	ID          string                 `json:"id"`           // 结果ID
	TaskID      string                 `json:"task_id"`      // 任务ID
	AgentID     string                 `json:"agent_id"`     // Agent ID
	Status      ResultStatus           `json:"status"`       // 结果状态
	Data        map[string]interface{} `json:"data"`         // 结果数据
	Metadata    map[string]interface{} `json:"metadata"`     // 元数据
	Error       string                 `json:"error"`        // 错误信息
	CreatedAt   time.Time              `json:"created_at"`   // 创建时间
	ValidatedAt *time.Time             `json:"validated_at"` // 验证时间
	Score       float64                `json:"score"`        // 结果评分
}

// AggregatedResult 聚合后的最终结果
type AggregatedResult struct {
	TaskID      string                 `json:"task_id"`       // 任务ID
	Results     []*TaskResult          `json:"results"`       // 所有结果
	MergedData  map[string]interface{} `json:"merged_data"`   // 合并后的数据
	Conflicts   []*Conflict            `json:"conflicts"`     // 冲突列表
	Strategy    string                 `json:"strategy"`      // 使用的聚合策略
	Confidence  float64                `json:"confidence"`    // 置信度
	CreatedAt   time.Time              `json:"created_at"`    // 创建时间
	CompletedAt *time.Time             `json:"completed_at"`  // 完成时间
}

// Conflict 结果冲突
type Conflict struct {
	Field       string        `json:"field"`        // 冲突字段
	Values      []interface{} `json:"values"`       // 冲突值
	AgentIDs    []string      `json:"agent_ids"`    // 涉及的Agent
	Resolution  string        `json:"resolution"`   // 解决方案
	ResolvedAt  *time.Time    `json:"resolved_at"`  // 解决时间
	Description string        `json:"description"`  // 冲突描述
}

// ResultStore 结果存储
type ResultStore struct {
	results map[string]*TaskResult // 按结果ID索引
	byTask  map[string][]*TaskResult // 按任务ID索引
	mu      sync.RWMutex
}

// NewResultStore 创建结果存储
func NewResultStore() *ResultStore {
	return &ResultStore{
		results: make(map[string]*TaskResult),
		byTask:  make(map[string][]*TaskResult),
	}
}

// AddResult 添加结果
func (s *ResultStore) AddResult(result *TaskResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.results[result.ID]; exists {
		return fmt.Errorf("result with ID %s already exists", result.ID)
	}

	s.results[result.ID] = result
	s.byTask[result.TaskID] = append(s.byTask[result.TaskID], result)

	return nil
}

// GetResult 获取结果
func (s *ResultStore) GetResult(resultID string) (*TaskResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, exists := s.results[resultID]
	if !exists {
		return nil, fmt.Errorf("result %s not found", resultID)
	}

	return result, nil
}

// GetResultsByTask 获取任务的所有结果
func (s *ResultStore) GetResultsByTask(taskID string) []*TaskResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := s.byTask[taskID]
	copied := make([]*TaskResult, len(results))
	copy(copied, results)

	return copied
}

// UpdateResult 更新结果
func (s *ResultStore) UpdateResult(result *TaskResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.results[result.ID]; !exists {
		return fmt.Errorf("result %s not found", result.ID)
	}

	s.results[result.ID] = result

	// 更新byTask索引
	for i, r := range s.byTask[result.TaskID] {
		if r.ID == result.ID {
			s.byTask[result.TaskID][i] = result
			break
		}
	}

	return nil
}

// DeleteResult 删除结果
func (s *ResultStore) DeleteResult(resultID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	result, exists := s.results[resultID]
	if !exists {
		return fmt.Errorf("result %s not found", resultID)
	}

	delete(s.results, resultID)

	// 从byTask索引中删除
	taskResults := s.byTask[result.TaskID]
	for i, r := range taskResults {
		if r.ID == resultID {
			s.byTask[result.TaskID] = append(taskResults[:i], taskResults[i+1:]...)
			break
		}
	}

	return nil
}

// GetAllResults 获取所有结果
func (s *ResultStore) GetAllResults() []*TaskResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*TaskResult, 0, len(s.results))
	for _, result := range s.results {
		results = append(results, result)
	}

	return results
}

// GetResultCount 获取结果数量
func (s *ResultStore) GetResultCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.results)
}

// GetResultCountByTask 获取任务的结果数量
func (s *ResultStore) GetResultCountByTask(taskID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.byTask[taskID])
}

// GetResultsByStatus 按状态获取结果
func (s *ResultStore) GetResultsByStatus(status ResultStatus) []*TaskResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*TaskResult, 0)
	for _, result := range s.results {
		if result.Status == status {
			results = append(results, result)
		}
	}

	return results
}

// SerializeResult 序列化结果
func SerializeResult(result *TaskResult) ([]byte, error) {
	return json.Marshal(result)
}

// DeserializeResult 反序列化结果
func DeserializeResult(data []byte) (*TaskResult, error) {
	var result TaskResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SerializeAggregatedResult 序列化聚合结果
func SerializeAggregatedResult(result *AggregatedResult) ([]byte, error) {
	return json.Marshal(result)
}

// DeserializeAggregatedResult 反序列化聚合结果
func DeserializeAggregatedResult(data []byte) (*AggregatedResult, error) {
	var result AggregatedResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
