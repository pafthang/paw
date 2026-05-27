package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/pafthang/paw/internal/config"
	"github.com/pafthang/paw/internal/health"
	"github.com/pafthang/paw/internal/server"
)

func Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return runServe(ctx, nil)
	}

	switch args[0] {
	case "serve":
		return runServe(ctx, args[1:])
	case "status":
		return runStatus(os.Stdout)
	case "doctor", "health":
		return runDoctor(ctx, os.Stdout, args[1:])
	case "config":
		return runConfig(os.Stdout, args[1:])
	case "help", "--help", "-h":
		printHelp(os.Stdout)
		return nil
	case "version", "--version", "-v":
		fmt.Fprintln(os.Stdout, "paw go-core-stage1")
		return nil
	default:
		printHelp(os.Stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runServe(ctx context.Context, args []string) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}

	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&settings.WebHost, "host", settings.WebHost, "host to bind")
	fs.IntVar(&settings.WebPort, "port", settings.WebPort, "port to bind")
	if err := fs.Parse(args); err != nil {
		return err
	}

	return server.New(settings).Run(ctx)
}

func runStatus(out io.Writer) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	payload := map[string]any{
		"status":        "ok",
		"implementation": "go",
		"stage":         "core-stage1",
		"config_dir":    must(config.Dir()),
		"config_path":   must(config.Path()),
		"web_host":      settings.WebHost,
		"web_port":      settings.WebPort,
		"agent_backend": settings.AgentBackend,
	}
	return json.NewEncoder(out).Encode(payload)
}

func runDoctor(ctx context.Context, out io.Writer, args []string) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	report := health.Run(ctx, settings)
	asJSON := contains(args, "--json")
	if asJSON {
		return json.NewEncoder(out).Encode(report)
	}
	fmt.Fprintf(out, "System: %s\n", strings.ToUpper(report.Status))
	for _, check := range report.Checks {
		fmt.Fprintf(out, "[%s] %s: %s\n", strings.ToUpper(check.Status), check.Name, check.Message)
	}
	return nil
}

func runConfig(out io.Writer, args []string) error {
	if len(args) == 0 || args[0] == "show" {
		settings, err := config.Load()
		if err != nil {
			return err
		}
		settings.OpenAIAPIKey = mask(settings.OpenAIAPIKey)
		settings.AnthropicAPIKey = mask(settings.AnthropicAPIKey)
		settings.TelegramBotToken = mask(settings.TelegramBotToken)
		return json.NewEncoder(out).Encode(settings)
	}

	switch args[0] {
	case "path":
		fmt.Fprintln(out, must(config.Path()))
		return nil
	case "dir":
		fmt.Fprintln(out, must(config.Dir()))
		return nil
	case "init":
		settings, err := config.Load()
		if err != nil {
			return err
		}
		if err := config.Save(settings); err != nil {
			return err
		}
		fmt.Fprintln(out, must(config.Path()))
		return nil
	case "set":
		if len(args) < 3 {
			return errors.New("usage: paw config set <key> <value>")
		}
		return configSet(out, args[1], strings.Join(args[2:], " "))
	default:
		return fmt.Errorf("unknown config command %q", args[0])
	}
}

func configSet(out io.Writer, key, value string) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	switch key {
	case "web_host":
		settings.WebHost = value
	case "web_port":
		var port int
		if _, err := fmt.Sscanf(value, "%d", &port); err != nil || port <= 0 {
			return fmt.Errorf("invalid web_port %q", value)
		}
		settings.WebPort = port
	case "agent_backend":
		settings.AgentBackend = value
	case "ollama_host":
		settings.OllamaHost = value
	case "openai_api_key":
		settings.OpenAIAPIKey = value
	case "anthropic_api_key":
		settings.AnthropicAPIKey = value
	case "telegram_bot_token":
		settings.TelegramBotToken = value
	default:
		keys := []string{"web_host", "web_port", "agent_backend", "ollama_host", "openai_api_key", "anthropic_api_key", "telegram_bot_token"}
		sort.Strings(keys)
		return fmt.Errorf("unknown config key %q; supported: %s", key, strings.Join(keys, ", "))
	}
	if err := config.Save(settings); err != nil {
		return err
	}
	fmt.Fprintf(out, "saved %s\n", key)
	return nil
}

func printHelp(out io.Writer) {
	fmt.Fprint(out, `paw - PocketPaw Go core stage 1

Usage:
  paw                         Start API server
  paw serve [--host H] [--port P]
  paw status                  Print local status as JSON
  paw doctor [--json]          Run basic health checks
  paw health [--json]          Alias for doctor
  paw config [show]            Print masked settings JSON
  paw config init              Create ~/.pocketpaw/config.json
  paw config path              Print config path
  paw config dir               Print config directory
  paw config set <key> <value> Save a supported config key
`)
}

func mask(value string) string {
	if value == "" {
		return ""
	}
	return "***"
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func must(value string, err error) string {
	if err != nil {
		return err.Error()
	}
	return value
}
