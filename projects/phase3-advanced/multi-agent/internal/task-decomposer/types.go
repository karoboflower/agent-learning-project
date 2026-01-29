package decomposer

import (
	"fmt"
	"time"
)

// Task 定义任务结构
type Task struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Input        interface{}            `json:"input"`
	Priority     int                    `json:"priority"`
	Dependencies []string               `json:"dependencies"`
	Requirements map[string]interface{} `json:"requirements"`
	Capabilities []string               `json:"capabilities"` // 需要的Agent能力
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
}

// SubTask 定义子任务结构
type SubTask struct {
	ID           string                 `json:"id"`
	ParentID     string                 `json:"parent_id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Input        interface{}            `json:"input"`
	Priority     int                    `json:"priority"`
	Dependencies []string               `json:"dependencies"`
	Requirements map[string]interface{} `json:"requirements"`
	Capabilities []string               `json:"capabilities"`
	Level        int                    `json:"level"` // 分解层级
	Metadata     map[string]interface{} `json:"metadata"`
}

// DecompositionResult 分解结果
type DecompositionResult struct {
	OriginalTask *Task                  `json:"original_task"`
	SubTasks     []*SubTask             `json:"sub_tasks"`
	Graph        *DependencyGraph       `json:"dependency_graph"`
	Strategy     string                 `json:"strategy"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// DependencyGraph 依赖关系图
type DependencyGraph struct {
	Nodes map[string]*GraphNode `json:"nodes"` // task_id -> node
	Edges []*Edge               `json:"edges"`
}

// GraphNode 图节点
type GraphNode struct {
	TaskID       string   `json:"task_id"`
	Level        int      `json:"level"`        // 拓扑层级
	Dependencies []string `json:"dependencies"` // 依赖的任务ID
	Dependents   []string `json:"dependents"`   // 依赖此任务的ID
}

// Edge 图的边
type Edge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Weight int    `json:"weight"` // 权重（用于优先级）
}

// DecompositionStrategy 分解策略
type DecompositionStrategy string

const (
	StrategyDependency  DecompositionStrategy = "DEPENDENCY"  // 基于依赖关系
	StrategyPriority    DecompositionStrategy = "PRIORITY"    // 基于优先级
	StrategyCapability  DecompositionStrategy = "CAPABILITY"  // 基于能力
	StrategyHybrid      DecompositionStrategy = "HYBRID"      // 混合策略
)

// TaskComplexity 任务复杂度
type TaskComplexity int

const (
	ComplexitySimple   TaskComplexity = 1 // 简单任务，不需要分解
	ComplexityModerate TaskComplexity = 2 // 中等复杂度
	ComplexityComplex  TaskComplexity = 3 // 复杂任务
	ComplexityVeryComplex TaskComplexity = 4 // 非常复杂
)

// DecomposerConfig 分解器配置
type DecomposerConfig struct {
	Strategy           DecompositionStrategy `json:"strategy"`
	MaxDepth           int                   `json:"max_depth"`            // 最大分解深度
	MinSubTasks        int                   `json:"min_sub_tasks"`        // 最小子任务数
	MaxSubTasks        int                   `json:"max_sub_tasks"`        // 最大子任务数
	ParallelThreshold  int                   `json:"parallel_threshold"`   // 并行阈值
	ComplexityAnalysis bool                  `json:"complexity_analysis"`  // 是否进行复杂度分析
}

// NewTask 创建新任务
func NewTask(id, taskType, description string) *Task {
	return &Task{
		ID:           id,
		Type:         taskType,
		Description:  description,
		Priority:     5, // 默认优先级
		Dependencies: make([]string, 0),
		Requirements: make(map[string]interface{}),
		Capabilities: make([]string, 0),
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
	}
}

// AddDependency 添加依赖
func (t *Task) AddDependency(taskID string) {
	t.Dependencies = append(t.Dependencies, taskID)
}

// AddCapability 添加所需能力
func (t *Task) AddCapability(capability string) {
	t.Capabilities = append(t.Capabilities, capability)
}

// SetRequirement 设置要求
func (t *Task) SetRequirement(key string, value interface{}) {
	t.Requirements[key] = value
}

// GetComplexity 获取任务复杂度
func (t *Task) GetComplexity() TaskComplexity {
	score := 0

	// 依赖数量影响复杂度
	if len(t.Dependencies) > 5 {
		score += 2
	} else if len(t.Dependencies) > 2 {
		score += 1
	}

	// 需要的能力数量
	if len(t.Capabilities) > 3 {
		score += 2
	} else if len(t.Capabilities) > 1 {
		score += 1
	}

	// 要求的复杂度
	if len(t.Requirements) > 5 {
		score += 1
	}

	// 根据评分返回复杂度
	if score >= 4 {
		return ComplexityVeryComplex
	} else if score >= 3 {
		return ComplexityComplex
	} else if score >= 1 {
		return ComplexityModerate
	}
	return ComplexitySimple
}

// IsDecomposable 判断任务是否可分解
func (t *Task) IsDecomposable() bool {
	complexity := t.GetComplexity()
	return complexity >= ComplexityModerate
}

// NewDependencyGraph 创建新的依赖图
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes: make(map[string]*GraphNode),
		Edges: make([]*Edge, 0),
	}
}

// AddNode 添加节点
func (g *DependencyGraph) AddNode(taskID string) {
	if _, exists := g.Nodes[taskID]; !exists {
		g.Nodes[taskID] = &GraphNode{
			TaskID:       taskID,
			Dependencies: make([]string, 0),
			Dependents:   make([]string, 0),
		}
	}
}

// AddEdge 添加边
func (g *DependencyGraph) AddEdge(from, to string, weight int) {
	// 确保节点存在
	g.AddNode(from)
	g.AddNode(to)

	// 添加边
	g.Edges = append(g.Edges, &Edge{
		From:   from,
		To:     to,
		Weight: weight,
	})

	// 更新节点关系
	g.Nodes[from].Dependents = append(g.Nodes[from].Dependents, to)
	g.Nodes[to].Dependencies = append(g.Nodes[to].Dependencies, from)
}

// HasCycle 检测是否有循环依赖
func (g *DependencyGraph) HasCycle() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeID := range g.Nodes {
		if !visited[nodeID] {
			if g.hasCycleDFS(nodeID, visited, recStack) {
				return true
			}
		}
	}

	return false
}

// hasCycleDFS DFS检测循环
func (g *DependencyGraph) hasCycleDFS(nodeID string, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	node := g.Nodes[nodeID]
	for _, depID := range node.Dependents {
		if !visited[depID] {
			if g.hasCycleDFS(depID, visited, recStack) {
				return true
			}
		} else if recStack[depID] {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}

// TopologicalSort 拓扑排序
func (g *DependencyGraph) TopologicalSort() ([]string, error) {
	// 检查循环依赖
	if g.HasCycle() {
		return nil, fmt.Errorf("circular dependency detected")
	}

	// 计算入度
	inDegree := make(map[string]int)
	for nodeID := range g.Nodes {
		inDegree[nodeID] = len(g.Nodes[nodeID].Dependencies)
	}

	// 找到所有入度为0的节点
	queue := make([]string, 0)
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	// 拓扑排序
	result := make([]string, 0)
	for len(queue) > 0 {
		// 取出第一个节点
		nodeID := queue[0]
		queue = queue[1:]
		result = append(result, nodeID)

		// 减少依赖此节点的其他节点的入度
		node := g.Nodes[nodeID]
		for _, depID := range node.Dependents {
			inDegree[depID]--
			if inDegree[depID] == 0 {
				queue = append(queue, depID)
			}
		}
	}

	// 检查是否所有节点都被访问
	if len(result) != len(g.Nodes) {
		return nil, fmt.Errorf("failed to sort all nodes")
	}

	return result, nil
}

// GetLevel 获取节点的拓扑层级
func (g *DependencyGraph) GetLevel(taskID string) int {
	if node, exists := g.Nodes[taskID]; exists {
		return node.Level
	}
	return 0
}

// CalculateLevels 计算所有节点的层级
func (g *DependencyGraph) CalculateLevels() error {
	// 拓扑排序
	sorted, err := g.TopologicalSort()
	if err != nil {
		return err
	}

	// 计算层级
	for _, taskID := range sorted {
		node := g.Nodes[taskID]
		maxLevel := 0

		// 找到所有依赖的最大层级
		for _, depID := range node.Dependencies {
			depNode := g.Nodes[depID]
			if depNode.Level >= maxLevel {
				maxLevel = depNode.Level + 1
			}
		}

		node.Level = maxLevel
	}

	return nil
}

// GetParallelTasks 获取可并行执行的任务组
func (g *DependencyGraph) GetParallelTasks() [][]string {
	// 按层级分组
	levels := make(map[int][]string)
	for taskID, node := range g.Nodes {
		level := node.Level
		if _, exists := levels[level]; !exists {
			levels[level] = make([]string, 0)
		}
		levels[level] = append(levels[level], taskID)
	}

	// 转换为切片
	result := make([][]string, 0)
	for i := 0; ; i++ {
		if tasks, exists := levels[i]; exists {
			result = append(result, tasks)
		} else {
			break
		}
	}

	return result
}
