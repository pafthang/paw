package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/memory"
)

type memoryPanelEntry struct {
	ID        string   `json:"id"`
	Content   string   `json:"content"`
	Timestamp string   `json:"timestamp"`
	Tags      []string `json:"tags"`
}

func (s *Server) handleListMemory(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	items, err := memory.List(database, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

func (s *Server) handleLongTermMemory(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	items, err := memory.List(database, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	out := make([]memoryPanelEntry, 0, len(items))
	for _, item := range items {
		if item.Type != "long_term" && item.Type != "fact" && item.Type != "preference" && item.Type != "" {
			continue
		}
		out = append(out, toMemoryPanelEntry(item))
	}
	return c.JSON(http.StatusOK, out)
}

func (s *Server) handleDeleteLongTermMemory(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if err := memory.Delete(database, uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"deleted": true, "id": id})
}

func (s *Server) handleMemorySettings(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"memory_backend":          "sqlite",
		"memory_use_inference":   false,
		"mem0_llm_provider":      "ollama",
		"mem0_llm_model":         s.settings.Model,
		"mem0_embedder_provider": "ollama",
		"mem0_embedder_model":    "nomic-embed-text",
		"mem0_vector_store":      "sqlite",
		"mem0_ollama_base_url":   s.settings.OllamaHost,
		"mem0_auto_learn":        false,
	})
}

func (s *Server) handleSaveMemorySettings(c echo.Context) error {
	var body map[string]any
	_ = c.Bind(&body)
	return c.JSON(http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleMemoryStats(c echo.Context) error {
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	items, err := memory.List(database, 100000)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	byType := map[string]int{}
	for _, item := range items {
		t := item.Type
		if t == "" {
			t = "long_term"
		}
		byType[t]++
	}
	return c.JSON(http.StatusOK, map[string]any{
		"backend":          "sqlite",
		"total_memories":   len(items),
		"memories_by_type": byType,
	})
}

func (s *Server) handleCreateMemory(c echo.Context) error {
	var req struct {
		Type     string `json:"type"`
		Content  string `json:"content"`
		Metadata any    `json:"metadata,omitempty"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
	}
	metadata := ""
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid metadata: " + err.Error()})
		}
		metadata = string(metadataBytes)
	}
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	item, err := memory.Add(database, req.Type, req.Content, metadata)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, item)
}

func (s *Server) handleGetMemory(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	item, err := memory.Get(database, uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, item)
}

func (s *Server) handleDeleteMemory(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if err := memory.Delete(database, uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"deleted": true, "id": id})
}

func (s *Server) handleSearchMemory(c echo.Context) error {
	q := c.QueryParam("q")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	items, err := memory.Search(database, q, limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

func toMemoryPanelEntry(item db.MemoryItem) memoryPanelEntry {
	tags := []string{}
	if strings.TrimSpace(item.Metadata) != "" {
		var meta map[string]any
		if json.Unmarshal([]byte(item.Metadata), &meta) == nil {
			if rawTags, ok := meta["tags"].([]any); ok {
				for _, raw := range rawTags {
					if tag, ok := raw.(string); ok && tag != "" {
						tags = append(tags, tag)
					}
				}
			}
		}
	}
	return memoryPanelEntry{
		ID:        strconv.FormatUint(uint64(item.ID), 10),
		Content:   item.Content,
		Timestamp: item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Tags:      tags,
	}
}
