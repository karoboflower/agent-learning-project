package handlers

import (
	"net/http"

	"github.com/agent-learning/go-agent-api/internal/agent"
	"github.com/agent-learning/go-agent-api/internal/scheduler"
	"github.com/gin-gonic/gin"
)

// TaskHandler handles task-related requests
type TaskHandler struct {
	scheduler *scheduler.Scheduler
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(scheduler *scheduler.Scheduler) *TaskHandler {
	return &TaskHandler{
		scheduler: scheduler,
	}
}

// SubmitTask godoc
// @Summary Submit a new task
// @Description Submit a new task for execution by an agent
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body agent.CreateTaskRequest true "Task creation request"
// @Success 201 {object} agent.Task
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/tasks [post]
func (h *TaskHandler) SubmitTask(c *gin.Context) {
	var req agent.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	task, err := h.scheduler.SubmitTask(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GetTask godoc
// @Summary Get task by ID
// @Description Get task details and status by task ID
// @Tags tasks
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} agent.Task
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")

	task, err := h.scheduler.GetTask(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// GetTaskResult godoc
// @Summary Get task result
// @Description Get the result of a completed task
// @Tags tasks
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} agent.TaskResult
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/tasks/{id}/result [get]
func (h *TaskHandler) GetTaskResult(c *gin.Context) {
	id := c.Param("id")

	result, err := h.scheduler.GetTaskResult(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CancelTask godoc
// @Summary Cancel a task
// @Description Cancel a pending or running task
// @Tags tasks
// @Param id path string true "Task ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/tasks/{id} [delete]
func (h *TaskHandler) CancelTask(c *gin.Context) {
	id := c.Param("id")

	if err := h.scheduler.CancelTask(id); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListTasks godoc
// @Summary List all tasks
// @Description Get a list of all tasks
// @Tags tasks
// @Produce json
// @Success 200 {object} TasksResponse
// @Router /api/v1/tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	tasks := h.scheduler.ListTasks()

	c.JSON(http.StatusOK, TasksResponse{
		Tasks: tasks,
		Total: len(tasks),
	})
}

// GetStats godoc
// @Summary Get scheduler statistics
// @Description Get statistics about pending, running, and completed tasks
// @Tags tasks
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tasks/stats [get]
func (h *TaskHandler) GetStats(c *gin.Context) {
	stats := h.scheduler.GetStats()
	c.JSON(http.StatusOK, stats)
}

// TasksResponse represents the response for listing tasks
type TasksResponse struct {
	Tasks []*agent.Task `json:"tasks"`
	Total int           `json:"total"`
}
