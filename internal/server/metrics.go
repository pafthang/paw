package server

import (
	"net/http"
	"os"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pafthang/paw/internal/db"
)

var metricsStartedAt = time.Now()

type systemMetricsCPU struct {
	Percent float64 `json:"percent"`
	Cores   int     `json:"cores"`
	FreqMHz *float64 `json:"freq_mhz"`
}

type systemMetricsMemory struct {
	UsedBytes  uint64  `json:"used_bytes"`
	TotalBytes uint64  `json:"total_bytes"`
	Percent    float64 `json:"percent"`
}

type systemMetricsDisk struct {
	UsedBytes  uint64  `json:"used_bytes"`
	TotalBytes uint64  `json:"total_bytes"`
	Percent    float64 `json:"percent"`
}

type systemMetricsBattery struct {
	Percent  int   `json:"percent"`
	Plugged  bool  `json:"plugged"`
	SecsLeft *int  `json:"secs_left"`
}

type systemMetricsResponse struct {
	Available     bool                   `json:"available"`
	OS            string                 `json:"os"`
	Arch          string                 `json:"arch"`
	CPU           systemMetricsCPU       `json:"cpu"`
	Memory        systemMetricsMemory    `json:"memory"`
	Disk          systemMetricsDisk      `json:"disk"`
	UptimeSeconds int64                  `json:"uptime_seconds"`
	Battery       *systemMetricsBattery  `json:"battery"`
	Timestamp     string                 `json:"timestamp"`
	Error         string                 `json:"error,omitempty"`
}

func (s *Server) handleSystemMetrics(c echo.Context) error {
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	disk := readDiskMetrics("/")
	resp := systemMetricsResponse{
		Available: true,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		CPU: systemMetricsCPU{
			Percent: 0,
			Cores:   runtime.NumCPU(),
			FreqMHz: nil,
		},
		Memory: systemMetricsMemory{
			UsedBytes:  mem.Alloc,
			TotalBytes: mem.Sys,
			Percent:    percent(mem.Alloc, mem.Sys),
		},
		Disk:          disk,
		UptimeSeconds: int64(time.Since(metricsStartedAt).Seconds()),
		Battery:       nil,
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) handleUsageSummary(c echo.Context) error {
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	var requestCount int64
	_ = database.Model(&db.ChatMessage{}).Where("role = ?", "assistant").Count(&requestCount).Error
	modelStats := map[string]map[string]any{}
	var rows []struct {
		Model string
		Count int64
	}
	_ = database.Model(&db.ChatMessage{}).
		Select("model, count(*) as count").
		Where("role = ?", "assistant").
		Group("model").
		Scan(&rows).Error
	for _, row := range rows {
		model := row.Model
		if model == "" {
			model = "unknown"
		}
		modelStats[model] = map[string]any{
			"input_tokens":  0,
			"output_tokens": 0,
			"cost_usd":      0,
			"count":         row.Count,
		}
	}
	backend := s.settings.AgentBackend
	if backend == "" {
		backend = "unknown"
	}
	return c.JSON(http.StatusOK, map[string]any{
		"total_input_tokens":        0,
		"total_output_tokens":       0,
		"total_cached_input_tokens": 0,
		"total_tokens":              0,
		"total_cost_usd":            0,
		"request_count":             requestCount,
		"by_model":                  modelStats,
		"by_backend": map[string]any{
			backend: map[string]any{
				"input_tokens":  0,
				"output_tokens": 0,
				"cost_usd":      0,
				"count":         requestCount,
			},
		},
	})
}

func (s *Server) handleRecentUsage(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 50
	}
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	var messages []db.ChatMessage
	if err := database.
		Where("role = ?", "assistant").
		Order("created_at desc, id desc").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	backend := s.settings.AgentBackend
	if backend == "" {
		backend = "unknown"
	}
	out := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		model := message.Model
		if model == "" {
			model = "unknown"
		}
		out = append(out, map[string]any{
			"timestamp":            message.CreatedAt.UTC().Format(time.RFC3339Nano),
			"backend":              backend,
			"model":                model,
			"input_tokens":         0,
			"output_tokens":        0,
			"cached_input_tokens":  0,
			"total_tokens":         0,
			"cost_usd":             nil,
			"session_id":           strconv.FormatUint(uint64(message.ChatSessionID), 10),
		})
	}
	return c.JSON(http.StatusOK, out)
}

func (s *Server) handleClearUsage(c echo.Context) error {
	// Usage in the current Go core is derived from chat messages rather than a
	// separate token-usage table, so clearing usage is intentionally a no-op.
	return c.NoContent(http.StatusNoContent)
}

func readDiskMetrics(path string) systemMetricsDisk {
	if path == "" {
		path = "/"
	}
	if _, err := os.Stat(path); err != nil {
		path = "/"
	}
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return systemMetricsDisk{}
	}
	total := uint64(st.Blocks) * uint64(st.Bsize)
	free := uint64(st.Bavail) * uint64(st.Bsize)
	used := total - free
	return systemMetricsDisk{UsedBytes: used, TotalBytes: total, Percent: percent(used, total)}
}

func percent(used, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(used) * 100 / float64(total)
}
