package health

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pafthang/paw/internal/config"
)

type Check struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Report struct {
	Status string  `json:"status"`
	Checks []Check `json:"checks"`
}

func Run(ctx context.Context, settings config.Settings) Report {
	checks := []Check{
		checkConfig(),
		checkPort(settings.WebHost, settings.WebPort),
		checkOllama(ctx, settings.OllamaHost),
	}
	status := "healthy"
	for _, check := range checks {
		if check.Status == "fail" {
			status = "unhealthy"
			break
		}
		if check.Status == "warn" && status == "healthy" {
			status = "degraded"
		}
	}
	return Report{Status: status, Checks: checks}
}

func checkConfig() Check {
	path, err := config.Path()
	if err != nil {
		return Check{Name: "config", Status: "fail", Message: err.Error()}
	}
	return Check{Name: "config", Status: "ok", Message: path}
}

func checkPort(host string, port int) Check {
	addr := net.JoinHostPort(host, intString(port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return Check{Name: "web_port", Status: "warn", Message: "port appears busy: " + addr}
	}
	_ = ln.Close()
	return Check{Name: "web_port", Status: "ok", Message: "available: " + addr}
}

func checkOllama(ctx context.Context, baseURL string) Check {
	if baseURL == "" {
		return Check{Name: "ollama", Status: "warn", Message: "ollama_host is empty"}
	}
	url := strings.TrimRight(baseURL, "/") + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Check{Name: "ollama", Status: "warn", Message: err.Error()}
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Check{Name: "ollama", Status: "warn", Message: "not reachable at " + baseURL}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return Check{Name: "ollama", Status: "warn", Message: resp.Status}
	}
	return Check{Name: "ollama", Status: "ok", Message: "reachable at " + baseURL}
}

func intString(v int) string {
	if v == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	neg := v < 0
	if neg {
		v = -v
	}
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
