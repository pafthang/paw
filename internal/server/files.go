package server

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/filestore"
)

func (s *Server) handleListFiles(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	items, err := filestore.List(database, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

func (s *Server) handleRecentFiles(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	items, err := filestore.List(database, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		ext := strings.TrimPrefix(filepath.Ext(item.Name), ".")
		out = append(out, map[string]any{
			"path":      item.Path,
			"name":      item.Name,
			"is_dir":    false,
			"extension": ext,
			"timestamp": item.UpdatedAt.Unix(),
			"tool":      "file-store",
		})
	}
	return c.JSON(http.StatusOK, out)
}

func (s *Server) handleCreateFile(c echo.Context) error {
	var req struct {
		Path     string `json:"path"`
		Metadata any    `json:"metadata,omitempty"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
	}
	metadata := ""
	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid metadata: " + err.Error()})
		}
		metadata = string(b)
	}
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	item, err := filestore.AddFromPath(database, req.Path, metadata)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, item)
}

func (s *Server) handleGetFile(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	item, err := filestore.Get(database, uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, item)
}

func (s *Server) handleDeleteFile(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if err := filestore.Delete(database, uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"deleted": true, "id": id})
}

func (s *Server) handleSearchFiles(c echo.Context) error {
	q := c.QueryParam("q")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	items, err := filestore.Search(database, q, limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}
