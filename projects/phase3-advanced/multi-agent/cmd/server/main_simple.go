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

// SimpleServer 简化的服务器
type SimpleServer struct {
	wsServer   *communication.WebSocketServer
	registry   *scheduler.AgentRegistry
	aggregator *aggregator.ResultAggregator
	allocator  *scheduler.TaskAllocator
	taskQueue  *scheduler.TaskQueue
	tasks      map[string]*WebTask
}

// NewSimpleServer 创建服务器
func NewSimpleServer() *SimpleServer {
	// WebSocket配置
	wsConfig := communication.DefaultWebSocketConfig()
	wsConfig.Port = 8080
	wsServer := communication.NewWebSocketServer(wsConfig)

	// Agent注册表
	registry := scheduler.NewAgentRegistry()

	// 结果聚合器
	agg := aggregator.NewResultAggregator(
		aggregator.MergeStrategyVoting,
		aggregator.ConflictResolutionVoting,
	)

	// 任务队列和分配器
	taskQueue := scheduler.NewTaskQueue(100)
	allocator := scheduler.NewTaskAllocator(registry, scheduler.StrategyLoadBalance)

	return &SimpleServer{
		wsServer:   wsServer,
		registry:   registry,
		aggregator: agg,
		allocator:  allocator,
		taskQueue:  taskQueue,
		tasks:      make(map[string]*WebTask),
	}
}

// Start 启动服务器
func (s *SimpleServer) Start() error {
	// 注册WebSocket消息处理器
	s.wsServer.RegisterMessageHandler("AGENT_REGISTER", s.handleAgentRegister)
	s.wsServer.RegisterMessageHandler("HEARTBEAT", s.handleHeartbeat)
	s.wsServer.RegisterMessageHandler("TASK_RESULT", s.handleTaskResult)

	// 注册HTTP路由
	http.HandleFunc("/", s.serveIndex)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./web/css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./web/js"))))

	// API路由
	http.HandleFunc("/api/agents", s.handleAgentsAPI)
	http.HandleFunc("/api/agents/", s.handleAgentAPI)
	http.HandleFunc("/api/tasks", s.handleTasksAPI)
	http.HandleFunc("/api/tasks/", s.handleTaskAPI)
	http.HandleFunc("/api/results", s.handleResultsAPI)
	http.HandleFunc("/api/results/aggregate/", s.handleAggregateAPI)

	// 启动WebSocket服务器
	if err := s.wsServer.Start(); err != nil {
		return err
	}

	log.Println("🚀 Multi-Agent Server started")
	log.Println("   WebSocket: ws://localhost:8080/ws")
	log.Println("   Web UI: http://localhost:8080")
	log.Println("   API: http://localhost:8080/api")

	return nil
}

// Stop 停止服务器
func (s *SimpleServer) Stop() error {
	return s.wsServer.Stop()
}

func (s *SimpleServer) serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "./web/index.html")
}

func (s *SimpleServer) handleAgentsAPI(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
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

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(agent)
	}
}

func (s *SimpleServer) handleAgentAPI(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

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
	}
}

func (s *SimpleServer) handleTasksAPI(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	switch r.Method {
	case "GET":
		tasks := make([]*WebTask, 0, len(s.tasks))
		for _, task := range s.tasks {
			tasks = append(tasks, task)
		}
		json.NewEncoder(w).Encode(tasks)

	case "POST":
		var task WebTask
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		task.Status = string(scheduler.TaskStatusPending)
		task.CreatedAt = time.Now()
		s.tasks[task.ID] = &task

		// 尝试分配任务
		go s.tryAllocateTask(&task)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	}
}

func (s *SimpleServer) handleTaskAPI(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)

	taskID := r.URL.Path[len("/api/tasks/"):]

	switch r.Method {
	case "GET":
		task, ok := s.tasks[taskID]
		if !ok {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(task)
	}
}

func (s *SimpleServer) handleResultsAPI(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)

	if r.Method == "GET" {
		results := s.aggregator.GetStore().GetAllResults()
		json.NewEncoder(w).Encode(results)
	}
}

func (s *SimpleServer) handleAggregateAPI(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)

	taskID := r.URL.Path[len("/api/results/aggregate/"):]

	if r.Method == "GET" {
		aggregated, err := s.aggregator.AggregateTask(taskID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(aggregated)
	}
}

func (s *SimpleServer) handleAgentRegister(msg *communication.Message) error {
	log.Printf("Agent registration: %s", msg.From)

	agent := &scheduler.Agent{
		ID:            msg.From,
		Name:          fmt.Sprintf("Agent-%s", msg.From),
		Capabilities:  []string{},
		Status:        scheduler.AgentStatusIdle,
		MaxTasks:      5,
		RegisteredAt:  time.Now(),
		LastHeartbeat: time.Now(),
	}

	if name, ok := msg.Payload["name"].(string); ok {
		agent.Name = name
	}

	if err := s.registry.Register(agent); err != nil {
		log.Printf("Failed to register agent: %v", err)
	}

	return nil
}

func (s *SimpleServer) handleHeartbeat(msg *communication.Message) error {
	agent, err := s.registry.GetAgent(msg.From)
	if err != nil {
		return nil
	}

	agent.LastHeartbeat = time.Now()
	return nil
}

func (s *SimpleServer) handleTaskResult(msg *communication.Message) error {
	result := &aggregator.TaskResult{
		ID:      msg.MessageID,
		TaskID:  msg.Payload["task_id"].(string),
		AgentID: msg.From,
		Data:    msg.Payload["data"].(map[string]interface{}),
		Score:   msg.Payload["score"].(float64),
		CreatedAt: time.Now(),
	}

	if err := s.aggregator.AddResult(result); err != nil {
		log.Printf("Failed to add result: %v", err)
		return err
	}

	// 尝试聚合
	go s.tryAggregate(result.TaskID)

	return nil
}

func (s *SimpleServer) tryAllocateTask(task *WebTask) {
	schedulerTask := task.ToSchedulerTask()

	agentID, err := s.allocator.Allocate(schedulerTask)
	if err != nil {
		log.Printf("Failed to allocate task %s: %v", task.ID, err)
		return
	}

	task.AssignedTo = agentID
	task.Status = string(scheduler.TaskStatusRunning)

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

	log.Printf("Task %s allocated to %s", task.ID, agentID)
}

func (s *SimpleServer) tryAggregate(taskID string) {
	results := s.aggregator.GetResultsByTask(taskID)
	if len(results) >= 2 {
		aggregated, err := s.aggregator.AggregateTask(taskID)
		if err != nil {
			log.Printf("Aggregation failed: %v", err)
			return
		}

		log.Printf("Results aggregated for task %s, confidence: %.2f",
			taskID, aggregated.Confidence)
	}
}

func (s *SimpleServer) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	server := NewSimpleServer()

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	if err := server.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	log.Println("Server stopped")
}
