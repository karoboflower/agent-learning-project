package handlers

import (
	"net/http"

	"github.com/agent-learning/go-agent-api/internal/agent"
	"github.com/gin-gonic/gin"
)

// AgentHandler handles agent-related requests
type AgentHandler struct {
	service agent.AgentService
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(service agent.AgentService) *AgentHandler {
	return &AgentHandler{
		service: service,
	}
}

// CreateAgent godoc
// @Summary Create a new agent
// @Description Create a new agent with specified configuration
// @Tags agents
// @Accept json
// @Produce json
// @Param agent body agent.CreateAgentRequest true "Agent creation request"
// @Success 201 {object} agent.Agent
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/agents [post]
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var req agent.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	ag, err := h.service.CreateAgent(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ag)
}

// GetAgent godoc
// @Summary Get agent by ID
// @Description Get agent details by agent ID
// @Tags agents
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} agent.Agent
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/agents/{id} [get]
func (h *AgentHandler) GetAgent(c *gin.Context) {
	id := c.Param("id")

	ag, err := h.service.GetAgent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ag)
}

// ListAgents godoc
// @Summary List all agents
// @Description Get a list of all registered agents
// @Tags agents
// @Produce json
// @Success 200 {object} AgentsResponse
// @Router /api/v1/agents [get]
func (h *AgentHandler) ListAgents(c *gin.Context) {
	agents, err := h.service.ListAgents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, AgentsResponse{
		Agents: agents,
		Total:  len(agents),
	})
}

// DeleteAgent godoc
// @Summary Delete an agent
// @Description Delete an agent by ID
// @Tags agents
// @Param id path string true "Agent ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/agents/{id} [delete]
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteAgent(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AgentsResponse represents the response for listing agents
type AgentsResponse struct {
	Agents []*agent.Agent `json:"agents"`
	Total  int            `json:"total"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
