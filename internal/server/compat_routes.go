package server

func (s *Server) registerCompatRoutes() {
	e := s.echo

	e.POST("/api/v1/chat/stream", s.handleChatStream)
	e.POST("/api/v1/chat/stop", s.handleChatStop)

	e.GET("/api/v1/backends", s.handleListBackends)
	e.GET("/api/v1/backends/ollama-models", s.handleFetchOllamaModels)
	e.GET("/api/v1/version", s.handleVersion)

	e.GET("/api/v1/metrics/system", s.handleSystemMetrics)
	e.GET("/api/v1/metrics/usage", s.handleUsageSummary)
	e.GET("/api/v1/metrics/usage/recent", s.handleRecentUsage)
	e.DELETE("/api/v1/metrics/usage", s.handleClearUsage)

	e.POST("/api/v1/health/check", s.handleHealthCompat)
	e.GET("/api/v1/health/errors", s.handleHealthErrors)
	e.DELETE("/api/v1/health/errors", s.handleClearHealthErrors)

	e.GET("/api/v1/memory/settings", s.handleMemorySettings)
	e.POST("/api/v1/memory/settings", s.handleSaveMemorySettings)
	e.GET("/api/v1/memory/stats", s.handleMemoryStats)
	e.GET("/api/v1/memory/long_term", s.handleLongTermMemory)
	e.DELETE("/api/v1/memory/long_term/:id", s.handleDeleteLongTermMemory)

	e.GET("/api/v1/files/recent", s.handleRecentFiles)

	e.GET("/api/v1/skills/search", s.handleSearchSkills)

	e.GET("/api/v1/mcp/presets", s.handleMCPPresets)
	e.POST("/api/v1/mcp/presets/install", s.handleMCPInstallPreset)

	e.GET("/api/v1/identity", s.handleIdentityGet)
	e.PUT("/api/v1/identity", s.handleIdentityPut)

	e.GET("/api/v1/kits/catalog", s.handleKitsCatalog)
	e.POST("/api/v1/kits/catalog/:kit_id/install", s.handleKitsInstallCatalog)
	e.GET("/api/v1/kits", s.handleKitsList)
	e.GET("/api/v1/kits/:kit_id", s.handleKitsGet)
	e.POST("/api/v1/kits/install", s.handleKitsInstall)
	e.DELETE("/api/v1/kits/:kit_id", s.handleKitsRemove)
	e.POST("/api/v1/kits/:kit_id/activate", s.handleKitsActivate)
	e.GET("/api/v1/kits/:kit_id/data", s.handleKitsData)

	e.GET("/api/mission-control/agents", s.handleMCListAgents)
	e.POST("/api/mission-control/agents", s.handleMCCreateAgent)
	e.DELETE("/api/mission-control/agents/:agent_id", s.handleMCDeleteAgent)
	e.GET("/api/mission-control/tasks/running", s.handleMCRunningTasks)
	e.POST("/api/mission-control/tasks", s.handleMCCreateTask)
	e.POST("/api/mission-control/tasks/:task_id/status", s.handleMCUpdateTaskStatus)
	e.DELETE("/api/mission-control/tasks/:task_id", s.handleMCDeleteTask)
	e.GET("/api/mission-control/tasks/:task_id/messages", s.handleMCListMessages)
	e.POST("/api/mission-control/tasks/:task_id/messages", s.handleMCPostMessage)
	e.GET("/api/mission-control/notifications", s.handleMCListNotifications)
	e.POST("/api/mission-control/notifications/:notification_id/read", s.handleMCMarkNotificationRead)
}
