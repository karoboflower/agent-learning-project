package api

import (
	"github.com/agent-learning/go-agent-api/internal/agent"
	"github.com/agent-learning/go-agent-api/internal/api/handlers"
	"github.com/agent-learning/go-agent-api/internal/api/middleware"
	"github.com/agent-learning/go-agent-api/internal/scheduler"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, agentService agent.AgentService, scheduler *scheduler.Scheduler) {
	// Apply middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(middleware.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "go-agent-api",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Agent routes
		agentHandler := handlers.NewAgentHandler(agentService)
		agents := v1.Group("/agents")
		{
			agents.POST("", agentHandler.CreateAgent)
			agents.GET("", agentHandler.ListAgents)
			agents.GET("/:id", agentHandler.GetAgent)
			agents.DELETE("/:id", agentHandler.DeleteAgent)
		}

		// Task routes
		taskHandler := handlers.NewTaskHandler(scheduler)
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", taskHandler.SubmitTask)
			tasks.GET("", taskHandler.ListTasks)
			tasks.GET("/stats", taskHandler.GetStats)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.GET("/:id/result", taskHandler.GetTaskResult)
			tasks.DELETE("/:id", taskHandler.CancelTask)
		}
	}

	// Swagger documentation (if enabled)
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
