package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agent-learning/multi-agent/internal/aggregator"
	"github.com/agent-learning/multi-agent/internal/communication"
	"github.com/agent-learning/multi-agent/internal/scheduler"
)

// Server 主服务器
type Server struct {
	wsServer    *communication.WebSocketServer
	registry    *scheduler.AgentRegistry
	taskManager *scheduler.TaskManager
	aggregator  *aggregator.ResultAggregator
	allocator   *scheduler.TaskAllocator
}

// NewServer 创建服务器
func NewServer() *Server {
	// 创建WebSocket服务器
	wsConfig := communication.DefaultWebSocketConfig()
	wsConfig.Port = 8080
	wsServer := communication.NewWebSocketServer(wsConfig)

	// 创建Agent注册表
	registry := scheduler.NewAgentRegistry()

	// 创建任务队列
	taskQueue := scheduler.NewTaskQueue(100)

	// 创建任务分配器
	allocator := scheduler.NewTaskAllocator(registry, scheduler.AllocationPolicyLoadBalance)

	// 创建任务管理器
	taskManager := scheduler.NewTaskManager(taskQueue, allocator)

	// 创建结果聚合器
	agg := aggregator.NewResultAggregator(
		aggregator.MergeStrategyVoting,
		aggregator.ConflictResolutionVoting,
	)

	// 配置验证规则
	validator := agg.GetValidator()
	validator.AddRule(&aggregator.RequiredFieldsRule{
		Fields: []string{},
	})
	validator.AddRule(&aggregator.ScoreRangeRule{
		MinScore: 0,
		MaxScore: 100,
	})

	return &Server{
		wsServer:    wsServer,
		registry:    registry,
		taskManager: taskManager,
		aggregator:  agg,
		allocator:   allocator,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	// 注册WebSocket消息处理器
	s.registerMessageHandlers()

	// 注册HTTP API路由
	s.registerHTTPHandlers()

	// 启动WebSocket服务器
	if err := s.wsServer.Start(); err != nil {
		return fmt.Errorf("failed to start WebSocket server: %w", err)
	}

	log.Println("🚀 Multi-Agent Server started")
	log.Printf("   WebSocket: ws://localhost:%d/ws", wsConfig.Port)
	log.Printf("   Web UI: http://localhost:%d", wsConfig.Port)
	log.Printf("   API: http://localhost:%d/api", wsConfig.Port)

	return nil
}

// Stop 停止服务器
func (s *Server) Stop() error {
	log.Println("Stopping server...")
	return s.wsServer.Stop()
}

// registerMessageHandlers 注册WebSocket消息处理器
func (s *Server) registerMessageHandlers() {
	// Agent注册
	s.wsServer.RegisterMessageHandler("AGENT_REGISTER", s.handleAgentRegister)

	// Agent心跳
	s.wsServer.RegisterMessageHandler("HEARTBEAT", s.handleHeartbeat)

	// 任务结果提交
	s.wsServer.RegisterMessageHandler("TASK_RESULT", s.handleTaskResult)

	// 任务状态更新
	s.wsServer.RegisterMessageHandler("TASK_STATUS", s.handleTaskStatus)
}

// registerHTTPHandlers 注册HTTP API处理器
func (s *Server) registerHTTPHandlers() {
	// 静态文件
	http.HandleFunc("/", s.handleIndex)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./web/css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./web/js"))))

	// API路由
	http.HandleFunc("/api/agents", s.handleAgentsAPI)
	http.HandleFunc("/api/agents/", s.handleAgentAPI)
	http.HandleFunc("/api/tasks", s.handleTasksAPI)
	http.HandleFunc("/api/tasks/", s.handleTaskAPI)
	http.HandleFunc("/api/results", s.handleResultsAPI)
	http.HandleFunc("/api/results/", s.handleResultAPI)
	http.HandleFunc("/api/results/aggregate/", s.handleAggregateResultAPI)
}

// handleIndex 首页处理
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "./web/index.html")
}

// handleAgentsAPI Agent列表API
func (s *Server) handleAgentsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		agents := s.registry.ListAgents()
		json.NewEncoder(w).Encode(agents)

	case "POST":
		var agent scheduler.Agent
		if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		agent.Status = scheduler.AgentStatusIdle
		agent.RegisteredAt = time.Now()
		agent.LastHeartbeat = time.Now()

		if err := s.registry.Register(&agent); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 广播Agent注册消息
		s.broadcastAgentUpdate("AGENT_REGISTERED", &agent)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(agent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAgentAPI 单个Agent API
func (s *Server) handleAgentAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 提取Agent ID
	agentID := r.URL.Path[len("/api/agents/"):]

	switch r.Method {
	case "GET":
		agent, err := s.registry.GetAgent(agentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(agent)

	case "DELETE":
		if err := s.registry.Unregister(agentID); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTasksAPI 任务列表API
func (s *Server) handleTasksAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		tasks := s.taskManager.ListTasks()
		json.NewEncoder(w).Encode(tasks)

	case "POST":
		var task scheduler.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		task.Status = string(scheduler.TaskStatusPending)
		task.CreatedAt = time.Now()

		if err := s.taskManager.AddTask(&task); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 尝试分配任务
		go s.tryAllocateTask(&task)

		// 广播任务创建消息
		s.broadcastTaskUpdate("TASK_CREATED", &task)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTaskAPI 单个任务API
func (s *Server) handleTaskAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 提取Task ID
	taskID := r.URL.Path[len("/api/tasks/"):]

	switch r.Method {
	case "GET":
		task := s.taskManager.GetTask(taskID)
		if task == nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(task)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleResultsAPI 结果列表API
func (s *Server) handleResultsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch r.Method {
	case "GET":
		results := s.aggregator.GetStore().GetAllResults()
		json.NewEncoder(w).Encode(results)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleResultAPI 单个结果API
func (s *Server) handleResultAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 提取Result ID
	resultID := r.URL.Path[len("/api/results/"):]

	switch r.Method {
	case "GET":
		result, err := s.aggregator.GetResult(resultID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(result)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAggregateResultAPI 聚合结果API
func (s *Server) handleAggregateResultAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 提取Task ID
	taskID := r.URL.Path[len("/api/results/aggregate/"):]

	switch r.Method {
	case "GET":
		aggregated, err := s.aggregator.AggregateTask(taskID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(aggregated)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAgentRegister 处理Agent注册消息
func (s *Server) handleAgentRegister(msg *communication.Message) error {
	log.Printf("Agent registration from: %s", msg.From)

	agent := &scheduler.Agent{
		ID:            msg.From,
		Name:          fmt.Sprintf("Agent-%s", msg.From),
		Capabilities:  []string{},
		Status:        scheduler.AgentStatusIdle,
		MaxTasks:      5,
		RegisteredAt:   time.Now(),
		LastHeartbeat: time.Now(),
	}

	// 从payload中提取信息
	if name, ok := msg.Payload["name"].(string); ok {
		agent.Name = name
	}
	if caps, ok := msg.Payload["capabilities"].([]interface{}); ok {
		for _, cap := range caps {
			if capStr, ok := cap.(string); ok {
				agent.Capabilities = append(agent.Capabilities, capStr)
			}
		}
	}
	if maxTasks, ok := msg.Payload["max_tasks"].(float64); ok {
		agent.MaxTasks = int(maxTasks)
	}

	if err := s.registry.Register(agent); err != nil {
		log.Printf("Failed to register agent: %v", err)
		return err
	}

	// 广播Agent注册事件
	s.broadcastAgentUpdate("AGENT_REGISTERED", agent)

	return nil
}

// handleHeartbeat 处理心跳消息
func (s *Server) handleHeartbeat(msg *communication.Message) error {
	agentID := msg.From

	agent, err := s.registry.GetAgent(agentID)
	if err != nil {
		log.Printf("Heartbeat from unknown agent: %s", agentID)
		return nil
	}

	// 更新心跳时间
	agent.LastHeartbeat = time.Now()

	// 更新状态
	if status, ok := msg.Payload["status"].(string); ok {
		agent.Status = scheduler.AgentStatus(status)
	}

	// 更新负载
	if load, ok := msg.Payload["load"].(float64); ok {
		agent.Load = load
	}

	s.registry.UpdateAgent(agent)

	return nil
}

// handleTaskResult 处理任务结果
func (s *Server) handleTaskResult(msg *communication.Message) error {
	log.Printf("Task result from %s", msg.From)

	// 构建TaskResult
	result := &aggregator.TaskResult{
		ID:        msg.MessageID,
		TaskID:    msg.Payload["task_id"].(string),
		AgentID:   msg.From,
		Data:      msg.Payload["data"].(map[string]interface{}),
		Score:     msg.Payload["score"].(float64),
		CreatedAt: time.Now(),
	}

	// 添加到聚合器
	if err := s.aggregator.AddResult(result); err != nil {
		log.Printf("Failed to add result: %v", err)
		return err
	}

	// 广播结果提交事件
	s.broadcastResultUpdate("RESULT_SUBMITTED", result)

	// 尝试聚合结果
	go s.tryAggregateResults(result.TaskID)

	return nil
}

// handleTaskStatus 处理任务状态更新
func (s *Server) handleTaskStatus(msg *communication.Message) error {
	taskID := msg.Payload["task_id"].(string)
	task := s.taskManager.GetTask(taskID)
	if task == nil {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// 更新状态
	if status, ok := msg.Payload["status"].(string); ok {
		task.Status = scheduler.TaskStatus(status)
	}

	// 更新进度
	if progress, ok := msg.Payload["progress"].(float64); ok {
		task.Progress = int(progress)
	}

	s.taskManager.UpdateTask(task)

	// 广播任务状态更新
	s.broadcastTaskUpdate("TASK_STATUS_UPDATE", task)

	return nil
}

// tryAllocateTask 尝试分配任务
func (s *Server) tryAllocateTask(task *scheduler.Task) {
	agentID, err := s.allocator.AllocateTask(task)
	if err != nil {
		log.Printf("Failed to allocate task %s: %v", task.ID, err)
		return
	}

	task.AssignedTo = agentID
	task.Status = scheduler.TaskStatusRunning
	s.taskManager.UpdateTask(task)

	// 发送任务给Agent
	msg := communication.NewMessageBuilder().
		SetType("TASK_REQUEST").
		SetFrom("server").
		SetTo(agentID).
		AddPayloadField("task_id", task.ID).
		AddPayloadField("task_type", task.Type).
		AddPayloadField("priority", task.Priority).
		AddPayloadField("description", task.Description).
		Build()

	s.wsServer.SendMessage(msg)

	// 广播任务分配事件
	s.broadcastTaskUpdate("TASK_ASSIGNED", task)

	log.Printf("Task %s allocated to agent %s", task.ID, agentID)
}

// tryAggregateResults 尝试聚合结果
func (s *Server) tryAggregateResults(taskID string) {
	// 获取任务的所有结果
	results := s.aggregator.GetResultsByTask(taskID)

	// 如果结果数量足够，进行聚合
	if len(results) >= 2 {
		aggregated, err := s.aggregator.AggregateTask(taskID)
		if err != nil {
			log.Printf("Failed to aggregate results for task %s: %v", taskID, err)
			return
		}

		log.Printf("Results aggregated for task %s, confidence: %.2f", taskID, aggregated.Confidence)

		// 广播聚合结果
		s.broadcastAggregatedResult(aggregated)
	}
}

// broadcastAgentUpdate 广播Agent更新
func (s *Server) broadcastAgentUpdate(eventType string, agent *scheduler.Agent) {
	msg := communication.NewMessageBuilder().
		SetType(eventType).
		SetFrom("server").
		SetTo("broadcast").
		AddPayloadField("agent_id", agent.ID).
		AddPayloadField("name", agent.Name).
		AddPayloadField("status", agent.Status).
		Build()

	s.wsServer.BroadcastMessage(msg)
}

// broadcastTaskUpdate 广播任务更新
func (s *Server) broadcastTaskUpdate(eventType string, task *scheduler.Task) {
	msg := communication.NewMessageBuilder().
		SetType(eventType).
		SetFrom("server").
		SetTo("broadcast").
		AddPayloadField("task_id", task.ID).
		AddPayloadField("status", task.Status).
		AddPayloadField("assigned_to", task.AssignedTo).
		Build()

	s.wsServer.BroadcastMessage(msg)
}

// broadcastResultUpdate 广播结果更新
func (s *Server) broadcastResultUpdate(eventType string, result *aggregator.TaskResult) {
	msg := communication.NewMessageBuilder().
		SetType(eventType).
		SetFrom("server").
		SetTo("broadcast").
		AddPayloadField("result_id", result.ID).
		AddPayloadField("task_id", result.TaskID).
		AddPayloadField("agent_id", result.AgentID).
		AddPayloadField("status", result.Status).
		Build()

	s.wsServer.BroadcastMessage(msg)
}

// broadcastAggregatedResult 广播聚合结果
func (s *Server) broadcastAggregatedResult(aggregated *aggregator.AggregatedResult) {
	msg := communication.NewMessageBuilder().
		SetType("RESULT_AGGREGATED").
		SetFrom("server").
		SetTo("broadcast").
		AddPayloadField("task_id", aggregated.TaskID).
		AddPayloadField("confidence", aggregated.Confidence).
		AddPayloadField("conflicts", len(aggregated.Conflicts)).
		Build()

	s.wsServer.BroadcastMessage(msg)
}

func main() {
	// 创建服务器
	server := NewServer()

	// 启动服务器
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	// 停止服务器
	if err := server.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	log.Println("Server stopped")
}
