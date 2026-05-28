package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pafthang/paw/internal/identity"
	"github.com/pafthang/paw/internal/kits"
	mc "github.com/pafthang/paw/internal/missioncontrol"
)

func (s *Server) handleIdentityGet(c echo.Context) error {
	files, err := identity.Load()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	return c.JSON(http.StatusOK, files)
}

func (s *Server) handleIdentityPut(c echo.Context) error {
	var req struct {
		IdentityFile     *string `json:"identity_file"`
		SoulFile         *string `json:"soul_file"`
		StyleFile        *string `json:"style_file"`
		InstructionsFile *string `json:"instructions_file"`
		UserFile         *string `json:"user_file"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"detail": "Invalid JSON body"})
	}
	resp, err := identity.Save(map[string]*string{
		"identity_file":     req.IdentityFile,
		"soul_file":         req.SoulFile,
		"style_file":        req.StyleFile,
		"instructions_file": req.InstructionsFile,
		"user_file":         req.UserFile,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}

func newKitStore(c echo.Context) (*kits.Store, error) {
	store, err := kits.NewStore()
	if err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Server) handleKitsList(c echo.Context) error {
	store, err := newKitStore(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	items, err := store.List()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"kits": items, "count": len(items)})
}

func (s *Server) handleKitsCatalog(c echo.Context) error {
	store, err := newKitStore(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	items, err := store.Catalog()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"catalog": items, "count": len(items)})
}

func (s *Server) handleKitsInstallCatalog(c echo.Context) error {
	store, err := newKitStore(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	kit, err := store.InstallCatalog(c.Param("kit_id"))
	if err != nil {
		return c.JSON(statusFromErr(err), map[string]string{"detail": err.Error()})
	}
	_ = store.Activate(kit.ID)
	kit.Active = true
	return c.JSON(http.StatusOK, map[string]any{"id": kit.ID, "kit": kit, "activated": true})
}

func (s *Server) handleKitsGet(c echo.Context) error {
	store, err := newKitStore(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	kit, err := store.Get(c.Param("kit_id"))
	if err != nil {
		return c.JSON(statusFromErr(err), map[string]string{"detail": "Kit not found"})
	}
	return c.JSON(http.StatusOK, map[string]any{"kit": kit})
}

func (s *Server) handleKitsInstall(c echo.Context) error {
	var req struct {
		YAML  string `json:"yaml"`
		KitID string `json:"kit_id"`
	}
	if err := c.Bind(&req); err != nil || req.YAML == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"detail": "Invalid kit configuration"})
	}
	store, err := newKitStore(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	kit, err := store.Install(req.YAML, req.KitID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"detail": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"id": kit.ID, "kit": kit})
}

func (s *Server) handleKitsRemove(c echo.Context) error {
	store, err := newKitStore(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	if err := store.Remove(c.Param("kit_id")); err != nil {
		return c.JSON(statusFromErr(err), map[string]string{"detail": "Kit not found"})
	}
	return c.JSON(http.StatusOK, map[string]any{"success": true, "message": "Kit removed"})
}

func (s *Server) handleKitsActivate(c echo.Context) error {
	store, err := newKitStore(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	id := c.Param("kit_id")
	if err := store.Activate(id); err != nil {
		return c.JSON(statusFromErr(err), map[string]string{"detail": "Kit not found"})
	}
	return c.JSON(http.StatusOK, map[string]any{"success": true, "message": "Kit activated"})
}

func (s *Server) handleKitsData(c echo.Context) error {
	store, err := newKitStore(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	data, err := store.Data(c.Param("kit_id"))
	if err != nil {
		return c.JSON(statusFromErr(err), map[string]string{"detail": err.Error()})
	}
	if stats, err := missionStoreStats(); err == nil {
		data["api:stats"] = stats
	}
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

func newMissionStore() (*mc.Store, error) {
	return mc.NewStore()
}

func missionStoreStats() (map[string]any, error) {
	store, err := newMissionStore()
	if err != nil {
		return nil, err
	}
	return store.Stats()
}

func (s *Server) handleMCListAgents(c echo.Context) error {
	store, err := newMissionStore()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	agents, err := store.ListAgents(c.QueryParam("status"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"agents": agents, "count": len(agents)})
}

func (s *Server) handleMCCreateAgent(c echo.Context) error {
	var agent mc.Agent
	if err := c.Bind(&agent); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"detail": "Invalid JSON body"})
	}
	store, err := newMissionStore()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	agent, err = store.SaveAgent(agent)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"agent": agent})
}

func (s *Server) handleMCDeleteAgent(c echo.Context) error {
	store, err := newMissionStore()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	if err := store.DeleteAgent(c.Param("agent_id")); err != nil {
		return c.JSON(statusFromErr(err), map[string]string{"detail": "Agent not found"})
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) handleMCRunningTasks(c echo.Context) error {
	store, err := newMissionStore()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	tasks, err := store.RunningTasks()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"running_tasks": tasks, "count": len(tasks)})
}

func (s *Server) handleMCCreateTask(c echo.Context) error {
	var task mc.Task
	if err := c.Bind(&task); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"detail": "Invalid JSON body"})
	}
	store, err := newMissionStore()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	task, err = store.SaveTask(task)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"task": task})
}

func (s *Server) handleMCUpdateTaskStatus(c echo.Context) error {
	var req struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&req); err != nil || req.Status == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"detail": "status is required"})
	}
	store, err := newMissionStore()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	tasks, err := store.ListTasks("")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	for _, task := range tasks {
		if task.ID == c.Param("task_id") {
			task.Status = req.Status
			task, err = store.SaveTask(task)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
			}
			return c.JSON(http.StatusOK, map[string]any{"task": task})
		}
	}
	return c.JSON(http.StatusNotFound, map[string]string{"detail": "Task not found"})
}

func (s *Server) handleMCDeleteTask(c echo.Context) error {
	store, err := newMissionStore()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	if err := store.DeleteTask(c.Param("task_id")); err != nil {
		return c.JSON(statusFromErr(err), map[string]string{"detail": "Task not found"})
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) handleMCListMessages(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	store, err := newMissionStore()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	messages, err := store.ListMessages(c.Param("task_id"), limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"messages": messages, "count": len(messages)})
}

func (s *Server) handleMCPostMessage(c echo.Context) error {
	var msg mc.Message
	if err := c.Bind(&msg); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"detail": "Invalid JSON body"})
	}
	msg.TaskID = c.Param("task_id")
	store, err := newMissionStore()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	msg, err = store.SaveMessage(msg)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"message": msg})
}

func (s *Server) handleMCListNotifications(c echo.Context) error {
	unreadOnly := c.QueryParam("unread_only") == "true"
	store, err := newMissionStore()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	notifications, err := store.ListNotifications(unreadOnly)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"notifications": notifications, "count": len(notifications)})
}

func (s *Server) handleMCMarkNotificationRead(c echo.Context) error {
	store, err := newMissionStore()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	if err := store.MarkNotificationRead(c.Param("notification_id")); err != nil {
		return c.JSON(statusFromErr(err), map[string]string{"detail": "Notification not found"})
	}
	return c.NoContent(http.StatusNoContent)
}

func statusFromErr(err error) int {
	if errors.Is(err, oos.ErrNotExist) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}
