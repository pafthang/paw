package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/pafthang/paw/internal/config"
	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/health"
	"github.com/pafthang/paw/internal/llm"
	"github.com/pafthang/paw/internal/server"
	"github.com/spf13/cobra"
)

func Run(ctx context.Context, args []string) error {
	root := newRootCommand(ctx, os.Stdout)
	root.SetArgs(args)
	return root.Execute()
}

func newRootCommand(ctx context.Context, out io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "paw",
		Short:         "PocketPaw Go core",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(ctx, cmd, args)
		},
	}
	root.SetOut(out)
	root.SetErr(os.Stderr)
	root.Version = "go-core-stage3"

	root.AddCommand(newServeCommand(ctx), newChatCommand(ctx, out), newStatusCommand(out), newDoctorCommand(ctx, out), newConfigCommand(out), newDBCommand(out))
	root.AddCommand(&cobra.Command{
		Use:     "ask [prompt]",
		Short:   "Alias for chat",
		Args:    cobra.MinimumNArgs(1),
		RunE:    func(cmd *cobra.Command, args []string) error { return runChat(ctx, out, cmd, args) },
		PreRunE: func(cmd *cobra.Command, args []string) error { return nil },
	})
	return root
}

func newServeCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the API server",
		RunE:  func(cmd *cobra.Command, args []string) error { return runServe(ctx, cmd, args) },
	}
	cmd.Flags().String("host", "", "host to bind")
	cmd.Flags().Int("port", 0, "port to bind")
	return cmd
}

func newChatCommand(ctx context.Context, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "chat [prompt]",
		Short:   "Send a prompt to the configured LLM",
		Args:    cobra.MinimumNArgs(1),
		RunE:    func(cmd *cobra.Command, args []string) error { return runChat(ctx, out, cmd, args) },
	}
	cmd.Flags().String("model", "", "model to use")
	cmd.Flags().Bool("json", false, "print JSON response")
	return cmd
}

func newStatusCommand(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Print local status as JSON",
		RunE:  func(cmd *cobra.Command, args []string) error { return runStatus(out) },
	}
}

func newDoctorCommand(ctx context.Context, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "doctor",
		Aliases: []string{"health"},
		Short:   "Run basic health checks",
		RunE:    func(cmd *cobra.Command, args []string) error { return runDoctor(ctx, out, cmd) },
	}
	cmd.Flags().Bool("json", false, "print JSON response")
	return cmd
}

func newConfigCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "config", Short: "Manage local config", RunE: func(cmd *cobra.Command, args []string) error { return runConfigShow(out) }}
	cmd.AddCommand(
		&cobra.Command{Use: "show", Short: "Print masked settings JSON", RunE: func(cmd *cobra.Command, args []string) error { return runConfigShow(out) }},
		&cobra.Command{Use: "init", Short: "Create ~/.pocketpaw/config.json", RunE: func(cmd *cobra.Command, args []string) error { return runConfigInit(out) }},
		&cobra.Command{Use: "path", Short: "Print config path", RunE: func(cmd *cobra.Command, args []string) error { fmt.Fprintln(out, must(config.Path())); return nil }},
		&cobra.Command{Use: "dir", Short: "Print config directory", RunE: func(cmd *cobra.Command, args []string) error { fmt.Fprintln(out, must(config.Dir())); return nil }},
		&cobra.Command{Use: "set <key> <value>", Short: "Save a supported config key", Args: cobra.MinimumNArgs(2), RunE: func(cmd *cobra.Command, args []string) error { return configSet(out, args[0], strings.Join(args[1:], " ")) }},
	)
	return cmd
}

func newDBCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "db", Short: "Manage local SQLite database"}
	cmd.AddCommand(
		&cobra.Command{Use: "path", Short: "Print SQLite database path", RunE: func(cmd *cobra.Command, args []string) error { fmt.Fprintln(out, must(config.DBPath())); return nil }},
		&cobra.Command{Use: "init", Short: "Open and migrate local SQLite database", RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := db.Open(); err != nil {
				return err
			}
			fmt.Fprintln(out, must(config.DBPath()))
			return nil
		}},
	)
	return cmd
}

func runServe(ctx context.Context, cmd *cobra.Command, args []string) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	if host, _ := cmd.Flags().GetString("host"); host != "" {
		settings.WebHost = host
	}
	if port, _ := cmd.Flags().GetInt("port"); port > 0 {
		settings.WebPort = port
	}
	return server.New(settings).Run(ctx)
}

func runChat(ctx context.Context, out io.Writer, cmd *cobra.Command, args []string) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	model, _ := cmd.Flags().GetString("model")
	if model == "" {
		model = llm.DefaultModel(settings)
	}
	jsonOut, _ := cmd.Flags().GetBool("json")
	prompt := strings.TrimSpace(strings.Join(args, " "))
	if prompt == "" {
		return errors.New("usage: paw chat [--model MODEL] <prompt>")
	}
	client, err := llm.NewClient(settings)
	if err != nil {
		return err
	}
	resp, err := client.Chat(ctx, llm.ChatRequest{Model: model, Messages: []llm.Message{{Role: "user", Content: prompt}}})
	if err != nil {
		return err
	}
	if jsonOut {
		return json.NewEncoder(out).Encode(resp)
	}
	fmt.Fprintln(out, resp.Content)
	return nil
}

func runStatus(out io.Writer) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	payload := map[string]any{
		"status":        "ok",
		"implementation": "go",
		"stage":         "core-stage3",
		"stack":         []string{"cobra", "echo", "gorm", "sqlite"},
		"config_dir":    must(config.Dir()),
		"config_path":   must(config.Path()),
		"db_path":       must(config.DBPath()),
		"web_host":      settings.WebHost,
		"web_port":      settings.WebPort,
		"agent_backend": settings.AgentBackend,
		"model":         settings.Model,
	}
	return json.NewEncoder(out).Encode(payload)
}

func runDoctor(ctx context.Context, out io.Writer, cmd *cobra.Command) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	report := health.Run(ctx, settings)
	asJSON, _ := cmd.Flags().GetBool("json")
	if asJSON {
		return json.NewEncoder(out).Encode(report)
	}
	fmt.Fprintf(out, "System: %s\n", strings.ToUpper(report.Status))
	for _, check := range report.Checks {
		fmt.Fprintf(out, "[%s] %s: %s\n", strings.ToUpper(check.Status), check.Name, check.Message)
	}
	return nil
}

func runConfigShow(out io.Writer) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	settings.OpenAIAPIKey = mask(settings.OpenAIAPIKey)
	settings.AnthropicAPIKey = mask(settings.AnthropicAPIKey)
	settings.TelegramBotToken = mask(settings.TelegramBotToken)
	return json.NewEncoder(out).Encode(settings)
}

func runConfigInit(out io.Writer) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	if err := config.Save(settings); err != nil {
		return err
	}
	fmt.Fprintln(out, must(config.Path()))
	return nil
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
	case "model":
		settings.Model = value
	case "ollama_host":
		settings.OllamaHost = value
	case "openai_compatible_base_url":
		settings.OpenAICompatibleBaseURL = value
	case "openai_api_key":
		settings.OpenAIAPIKey = value
	case "anthropic_api_key":
		settings.AnthropicAPIKey = value
	case "telegram_bot_token":
		settings.TelegramBotToken = value
	default:
		keys := []string{"web_host", "web_port", "agent_backend", "model", "ollama_host", "openai_compatible_base_url", "openai_api_key", "anthropic_api_key", "telegram_bot_token"}
		sort.Strings(keys)
		return fmt.Errorf("unknown config key %q; supported: %s", key, strings.Join(keys, ", "))
	}
	if err := config.Save(settings); err != nil {
		return err
	}
	fmt.Fprintf(out, "saved %s\n", key)
	return nil
}

func mask(value string) string {
	if value == "" {
		return ""
	}
	return "***"
}

func must(value string, err error) string {
	if err != nil {
		return err.Error()
	}
	return value
}
