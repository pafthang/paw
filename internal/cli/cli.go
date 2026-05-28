package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/pafthang/paw/internal/agent"
	"github.com/pafthang/paw/internal/channels"
	tg "github.com/pafthang/paw/internal/channels/telegram"
	"github.com/pafthang/paw/internal/config"
	"github.com/pafthang/paw/internal/contextpack"
	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/filestore"
	"github.com/pafthang/paw/internal/health"
	"github.com/pafthang/paw/internal/llm"
	"github.com/pafthang/paw/internal/mcp"
	"github.com/pafthang/paw/internal/memory"
	"github.com/pafthang/paw/internal/presets"
	"github.com/pafthang/paw/internal/search"
	"github.com/pafthang/paw/internal/server"
	"github.com/pafthang/paw/internal/skills"
	"github.com/pafthang/paw/internal/tools"
	"github.com/spf13/cobra"
)

func Run(ctx context.Context, args []string) error {
	root := newRootCommand(ctx, os.Stdout)
	root.SetArgs(args)
	return root.Execute()
}

func newRootCommand(ctx context.Context, out io.Writer) *cobra.Command {
	root := &cobra.Command{Use: "paw", Short: "PocketPaw Go core", SilenceUsage: true, SilenceErrors: true, RunE: func(cmd *cobra.Command, args []string) error { return runServe(ctx, cmd, args) }}
	root.SetOut(out)
	root.SetErr(os.Stderr)
	root.Version = "go-core-stage7-telegram"
	root.AddCommand(newServeCommand(ctx), newChatCommand(ctx, out), newAgentCommand(ctx, out), newStatusCommand(out), newDoctorCommand(ctx, out), newConfigCommand(out), newAuthCommand(out), newDBCommand(out), newSessionsCommand(out), newMemoryCommand(out), newFileStoreCommand(out), newSearchCommand(out), newSkillsCommand(out), newMCPCommand(out), newChannelsCommand(ctx, out), newToolsCommand(out), newRunToolCommand(ctx, out), newAuditCommand(out))
	root.AddCommand(&cobra.Command{Use: "ask [prompt]", Short: "Alias for chat", Args: cobra.MinimumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runChat(ctx, out, cmd, args) }})
	return root
}

func newServeCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{Use: "serve", Short: "Start the API server", RunE: func(cmd *cobra.Command, args []string) error { return runServe(ctx, cmd, args) }}
	cmd.Flags().String("host", "", "host to bind")
	cmd.Flags().Int("port", 0, "port to bind")
	return cmd
}

func newChatCommand(ctx context.Context, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "chat [prompt]", Short: "Send a prompt to the configured LLM and save the exchange", Args: cobra.MinimumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runChat(ctx, out, cmd, args) }}
	cmd.Flags().String("model", "", "model to use")
	cmd.Flags().Bool("json", false, "print JSON response")
	cmd.Flags().Uint("session", 0, "append to an existing session id")
	cmd.Flags().Int("history-limit", db.DefaultHistoryLimit, "max previous session messages to consider for LLM context")
	cmd.Flags().String("system", contextpack.DefaultSystemPrompt, "system prompt prepended to the LLM context")
	cmd.Flags().Int("max-context-chars", contextpack.DefaultMaxContextChars, "rough maximum chars for packed LLM context")
	return cmd
}

func newAgentCommand(ctx context.Context, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "agent [prompt]", Short: "Ask the LLM to decide whether tools are needed, then run them and produce a final answer", Args: cobra.MinimumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runAgentChat(ctx, out, cmd, args) }}
	cmd.Flags().String("model", "", "model to use")
	cmd.Flags().Bool("json", false, "print JSON response")
	cmd.Flags().Uint("session", 0, "append to an existing session id")
	cmd.Flags().Int("history-limit", db.DefaultHistoryLimit, "max previous session messages to consider for LLM context")
	cmd.Flags().String("system", "", "override agent tool-call system prompt")
	cmd.Flags().Int("max-context-chars", contextpack.DefaultMaxContextChars, "rough maximum chars for packed LLM context")
	cmd.Flags().Int("max-iterations", 4, "maximum tool/LLM iterations before stopping")
	cmd.Flags().Bool("allow-shell", false, "allow shell.run tool calls (still blocked for dangerous commands unless --allow-shell-dangerous)")
	cmd.Flags().Bool("allow-shell-dangerous", false, "allow dangerous shell commands (use with extreme caution)")
	cmd.Flags().String("workspace", "", "workspace root for file tools and shell working directory (default: current directory)")
	cmd.Flags().String("skill", "", "inject a named skill from ~/.pocketpaw/skills/<name>/skill.yaml")
	return cmd
}

func newStatusCommand(out io.Writer) *cobra.Command {
	return &cobra.Command{Use: "status", Short: "Print local status as JSON", RunE: func(cmd *cobra.Command, args []string) error { return runStatus(out) }}
}

func newDoctorCommand(ctx context.Context, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "doctor", Aliases: []string{"health"}, Short: "Run basic health checks", RunE: func(cmd *cobra.Command, args []string) error { return runDoctor(ctx, out, cmd) }}
	cmd.Flags().Bool("json", false, "print JSON response")
	return cmd
}

func newConfigCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "config", Short: "Manage local config", RunE: func(cmd *cobra.Command, args []string) error { return runConfigShow(out) }}
	cmd.AddCommand(&cobra.Command{Use: "show", Short: "Print masked settings JSON", RunE: func(cmd *cobra.Command, args []string) error { return runConfigShow(out) }}, &cobra.Command{Use: "init", Short: "Create ~/.pocketpaw/config.json", RunE: func(cmd *cobra.Command, args []string) error { return runConfigInit(out) }}, &cobra.Command{Use: "path", Short: "Print config path", RunE: func(cmd *cobra.Command, args []string) error { fmt.Fprintln(out, must(config.Path())); return nil }}, &cobra.Command{Use: "dir", Short: "Print config directory", RunE: func(cmd *cobra.Command, args []string) error { fmt.Fprintln(out, must(config.Dir())); return nil }}, &cobra.Command{Use: "set <key> <value>", Short: "Save a supported config key", Args: cobra.MinimumNArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		return configSet(out, args[0], strings.Join(args[1:], " "))
	}})
	return cmd
}

func newDBCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "db", Short: "Manage local SQLite database"}
	cmd.AddCommand(&cobra.Command{Use: "path", Short: "Print SQLite database path", RunE: func(cmd *cobra.Command, args []string) error { fmt.Fprintln(out, must(config.DBPath())); return nil }}, &cobra.Command{Use: "init", Short: "Open and migrate local SQLite database", RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := db.Open(); err != nil {
			return err
		}
		fmt.Fprintln(out, must(config.DBPath()))
		return nil
	}})
	return cmd
}

func newSessionsCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "sessions", Short: "Inspect saved chat sessions", RunE: func(cmd *cobra.Command, args []string) error { return runSessionsList(out, cmd) }}
	list := &cobra.Command{Use: "list", Short: "List saved sessions", RunE: func(cmd *cobra.Command, args []string) error { return runSessionsList(out, cmd) }}
	list.Flags().Int("limit", 20, "maximum sessions to show")
	show := &cobra.Command{Use: "show <id>", Short: "Show one saved session", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runSessionsShow(out, args[0]) }}
	search := &cobra.Command{Use: "search <query>", Short: "Search sessions by title and messages", Args: cobra.MinimumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		return runSessionsSearch(out, cmd, strings.Join(args, " "))
	}}
	search.Flags().Int("limit", 20, "maximum sessions to show")
	rename := &cobra.Command{Use: "rename <id> <title>", Short: "Rename a session title", Args: cobra.MinimumNArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		return runSessionsRename(out, args[0], strings.Join(args[1:], " "))
	}}
	deleteCmd := &cobra.Command{Use: "delete <id>", Short: "Delete one saved session", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runSessionsDelete(out, args[0]) }}
	cmd.AddCommand(list, show, search, rename, deleteCmd)
	cmd.Flags().Int("limit", 20, "maximum sessions to show")
	return cmd
}

func newMemoryCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "memory", Short: "Manage long-term memory items"}

	list := &cobra.Command{Use: "list", Aliases: []string{"ls"}, Short: "List recent memory items", RunE: func(cmd *cobra.Command, args []string) error { return runMemoryList(out, cmd) }}
	list.Flags().Int("limit", 50, "maximum items to show")
	list.Flags().Bool("json", true, "print JSON output")

	show := &cobra.Command{Use: "show <id>", Short: "Show one memory item", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runMemoryShow(out, args[0]) }}

	add := &cobra.Command{Use: "add <type> <content>", Short: "Add a memory item", Args: cobra.MinimumNArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		return runMemoryAdd(out, cmd, args[0], strings.Join(args[1:], " "))
	}}
	add.Flags().String("metadata", "", "optional metadata string (JSON or free text)")
	add.Flags().Bool("json", true, "print JSON output")

	del := &cobra.Command{Use: "delete <id>", Aliases: []string{"rm"}, Short: "Delete a memory item", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runMemoryDelete(out, args[0]) }}

	search := &cobra.Command{Use: "search <query>", Short: "Search memory items", Args: cobra.MinimumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		return runMemorySearch(out, cmd, strings.Join(args, " "))
	}}
	search.Flags().Int("limit", 50, "maximum items to show")
	search.Flags().Bool("json", true, "print JSON output")

	cmd.AddCommand(list, show, add, del, search)
	return cmd
}

func newFileStoreCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "file-store", Short: "Manage imported files stored under ~/.pocketpaw/files"}

	list := &cobra.Command{Use: "list", Aliases: []string{"ls"}, Short: "List stored files", RunE: func(cmd *cobra.Command, args []string) error { return runFileStoreList(out, cmd) }}
	list.Flags().Int("limit", 50, "maximum files to show")
	list.Flags().Bool("json", true, "print JSON output")

	add := &cobra.Command{Use: "add <path>", Short: "Import a local file into the file store", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runFileStoreAdd(out, cmd, args[0]) }}
	add.Flags().String("metadata", "", "optional metadata string (JSON or free text)")
	add.Flags().Bool("json", true, "print JSON output")

	show := &cobra.Command{Use: "show <id>", Short: "Show one stored file record", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runFileStoreShow(out, args[0]) }}

	del := &cobra.Command{Use: "delete <id>", Aliases: []string{"rm"}, Short: "Delete a stored file record and its stored content", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runFileStoreDelete(out, args[0]) }}

	search := &cobra.Command{Use: "search <query>", Short: "Search stored files", Args: cobra.MinimumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		return runFileStoreSearch(out, cmd, strings.Join(args, " "))
	}}
	search.Flags().Int("limit", 50, "maximum files to show")
	search.Flags().Bool("json", true, "print JSON output")

	cmd.AddCommand(list, add, show, del, search)
	return cmd
}

func newSearchCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "search <query>", Short: "Search across memory, sessions, messages, and file store", Args: cobra.MinimumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		return runSearch(out, cmd, strings.Join(args, " "))
	}}
	cmd.Flags().Int("limit", 50, "maximum results to show")
	cmd.Flags().Bool("json", true, "print JSON output")
	return cmd
}

func newSkillsCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "skills", Short: "Manage local skills"}

	list := &cobra.Command{Use: "list", Short: "List installed skills", RunE: func(cmd *cobra.Command, args []string) error { return runSkillsList(out) }}
	show := &cobra.Command{Use: "show <name>", Short: "Show one skill", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runSkillsShow(out, args[0]) }}
	validate := &cobra.Command{Use: "validate <path>", Short: "Validate a skill.yaml file", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runSkillsValidate(out, args[0]) }}
	reload := &cobra.Command{Use: "reload", Short: "Reload skills from disk", RunE: func(cmd *cobra.Command, args []string) error { return runSkillsReload(out) }}
	install := &cobra.Command{Use: "install <path>", Short: "Install a skill from a local directory", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runSkillsInstall(out, cmd, args[0]) }}
	install.Flags().Bool("force", false, "overwrite existing skill if present")
	uninstall := &cobra.Command{Use: "uninstall <name>", Short: "Uninstall a skill", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runSkillsUninstall(out, cmd, args[0]) }}
	uninstall.Flags().Bool("yes", false, "confirm uninstall")

	cmd.AddCommand(list, show, validate, reload, install, uninstall)
	return cmd
}

func newMCPCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "mcp", Short: "Manage MCP server configs and processes"}

	list := &cobra.Command{Use: "list", Short: "List configured MCP servers", RunE: func(cmd *cobra.Command, args []string) error { return runMCPList(out) }}
	show := &cobra.Command{Use: "show <name>", Short: "Show one MCP server config", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runMCPShow(out, args[0]) }}
	add := &cobra.Command{Use: "add <name>", Short: "Add an MCP server config", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runMCPAdd(out, cmd, args[0]) }}
	add.Flags().String("command", "", "command to run")
	add.Flags().StringArray("arg", nil, "repeatable argument")
	remove := &cobra.Command{Use: "remove <name>", Short: "Remove an MCP server config", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runMCPRemove(out, args[0]) }}
	start := &cobra.Command{Use: "start <name>", Short: "Start an MCP server process", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runMCPStart(out, args[0]) }}
	stop := &cobra.Command{Use: "stop <name>", Short: "Stop an MCP server process", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runMCPStop(out, args[0]) }}
	status := &cobra.Command{Use: "status", Short: "Show MCP process status", RunE: func(cmd *cobra.Command, args []string) error { return runMCPStatus(out) }}
	presetsCmd := &cobra.Command{Use: "presets", Short: "List built-in MCP presets", RunE: func(cmd *cobra.Command, args []string) error { return runMCPPresets(out) }}
	installPreset := &cobra.Command{Use: "install-preset <name>", Short: "Install a built-in preset into mcp.json", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runMCPInstallPreset(out, cmd, args[0]) }}
	installPreset.Flags().String("workspace", "", "workspace root for presets that require it")

	cmd.AddCommand(list, show, add, remove, start, stop, status, presetsCmd, installPreset)
	return cmd
}

func newChannelsCommand(ctx context.Context, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "channels", Short: "Manage external channels (adapters)"}
	settings, _ := config.Load()
	manager := channels.NewManager()
	manager.Register(tg.New(settings))

	list := &cobra.Command{Use: "list", Short: "List available channels", RunE: func(cmd *cobra.Command, args []string) error {
		return json.NewEncoder(out).Encode(map[string]any{"channels": manager.List()})
	}}
	status := &cobra.Command{Use: "status", Short: "Show channel status", RunE: func(cmd *cobra.Command, args []string) error {
		return json.NewEncoder(out).Encode(manager.StatusAll())
	}}
	start := &cobra.Command{Use: "start <name>", Short: "Start a channel", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		st, err := manager.Start(ctx, args[0])
		if err != nil {
			return err
		}
		return json.NewEncoder(out).Encode(st)
	}}
	stop := &cobra.Command{Use: "stop <name>", Short: "Stop a channel", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		st, err := manager.Stop(ctx, args[0])
		if err != nil {
			return err
		}
		return json.NewEncoder(out).Encode(st)
	}}

	cmd.AddCommand(list, status, start, stop)
	return cmd
}

func newToolsCommand(out io.Writer) *cobra.Command {
	return &cobra.Command{Use: "tools", Short: "List available agent tools", RunE: func(cmd *cobra.Command, args []string) error {
		return json.NewEncoder(out).Encode(tools.DefaultRegistry().List())
	}}
}

func newRunToolCommand(ctx context.Context, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "run-tool <name>", Short: "Run one agent tool and audit it", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		input, _ := cmd.Flags().GetString("input")
		sessionID, _ := cmd.Flags().GetUint("session")
		return runTool(ctx, out, cmd, args[0], input, sessionID)
	}}
	cmd.Flags().String("input", "{}", "tool input JSON")
	cmd.Flags().Uint("session", 0, "optional session id for audit")
	cmd.Flags().Bool("allow-shell", false, "allow shell.run tool calls")
	cmd.Flags().Bool("allow-shell-dangerous", false, "allow dangerous shell.run commands (use with extreme caution)")
	cmd.Flags().String("workspace", "", "workspace root for file tools and shell working directory (default: current directory)")
	return cmd
}

func newAuditCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "audit", Short: "Inspect audit events", RunE: func(cmd *cobra.Command, args []string) error { return runAuditList(out, cmd) }}
	list := &cobra.Command{Use: "list", Short: "List audit events", RunE: func(cmd *cobra.Command, args []string) error { return runAuditList(out, cmd) }}
	list.Flags().Int("limit", 50, "maximum audit events to show")
	cmd.Flags().Int("limit", 50, "maximum audit events to show")
	cmd.AddCommand(list)
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
	sessionID, _ := cmd.Flags().GetUint("session")
	historyLimit, _ := cmd.Flags().GetInt("history-limit")
	systemPrompt, _ := cmd.Flags().GetString("system")
	maxContextChars, _ := cmd.Flags().GetInt("max-context-chars")
	prompt := strings.TrimSpace(strings.Join(args, " "))
	if prompt == "" {
		return errors.New("usage: paw chat [--model MODEL] [--session ID] [--history-limit N] [--max-context-chars N] <prompt>")
	}
	database, err := db.Open()
	if err != nil {
		return err
	}
	var session *db.ChatSession
	var history []llm.Message
	incoming := []llm.Message{{Role: "user", Content: prompt}}
	if sessionID > 0 {
		session, err = db.GetChatSession(database, uint(sessionID))
		if err != nil {
			return err
		}
		recent, err := db.ListRecentChatMessages(database, session.ID, historyLimit)
		if err != nil {
			return err
		}
		history = append(history, toLLMMessages(recent)...)
	} else {
		session, err = db.CreateChatSession(database, prompt)
		if err != nil {
			return err
		}
	}
	messages := contextpack.Pack(systemPrompt, history, incoming, maxContextChars)
	client, err := llm.NewClient(settings)
	if err != nil {
		return err
	}
	resp, err := client.Chat(ctx, llm.ChatRequest{Model: model, Messages: messages})
	if err != nil {
		return err
	}
	if _, err := db.AddChatMessage(database, session.ID, "user", prompt, model); err != nil {
		return err
	}
	if _, err := db.AddChatMessage(database, session.ID, "assistant", resp.Content, resp.Model); err != nil {
		return err
	}
	stats := contextpack.Stats(messages)
	if jsonOut {
		return json.NewEncoder(out).Encode(map[string]any{"session_id": session.ID, "history_messages": len(messages) - len(incoming) - 1, "context": stats, "response": resp})
	}
	fmt.Fprintf(out, "%s\n\n[session:%d history:%d context:%v chars]\n", resp.Content, session.ID, len(messages)-len(incoming)-1, stats["chars"])
	return nil
}

func runAgentChat(ctx context.Context, out io.Writer, cmd *cobra.Command, args []string) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	model, _ := cmd.Flags().GetString("model")
	if model == "" {
		model = llm.DefaultModel(settings)
	}
	jsonOut, _ := cmd.Flags().GetBool("json")
	sessionID, _ := cmd.Flags().GetUint("session")
	historyLimit, _ := cmd.Flags().GetInt("history-limit")
	systemPrompt, _ := cmd.Flags().GetString("system")
	maxContextChars, _ := cmd.Flags().GetInt("max-context-chars")
	maxIterations, _ := cmd.Flags().GetInt("max-iterations")
	allowShell, _ := cmd.Flags().GetBool("allow-shell")
	allowShellDangerous, _ := cmd.Flags().GetBool("allow-shell-dangerous")
	workspace, _ := cmd.Flags().GetString("workspace")
	skill, _ := cmd.Flags().GetString("skill")
	prompt := strings.TrimSpace(strings.Join(args, " "))
	if prompt == "" {
		return errors.New("usage: paw agent [--model MODEL] [--session ID] <prompt>")
	}
	database, err := db.Open()
	if err != nil {
		return err
	}
	client, err := llm.NewClient(settings)
	if err != nil {
		return err
	}
	runner := agent.NewDefaultRunner(database)
	resp, err := runner.Chat(ctx, client, agent.ChatRequest{
		SessionID:           uint(sessionID),
		Prompt:              prompt,
		Model:               model,
		HistoryLimit:        historyLimit,
		MaxContextChars:     maxContextChars,
		SystemPrompt:        systemPrompt,
		MaxIterations:       maxIterations,
		Workspace:           workspace,
		AllowShell:          allowShell,
		AllowShellDangerous: allowShellDangerous,
		Skill:               skill,
	})
	if err != nil {
		return err
	}
	if jsonOut {
		return json.NewEncoder(out).Encode(resp)
	}
	if resp.UsedTools {
		fmt.Fprintln(out, resp.FinalResponse.Content)
	} else {
		fmt.Fprintln(out, resp.ModelResponse.Content)
	}
	if resp.UsedTools {
		fmt.Fprintf(out, "\n[session:%d used_tools:true iterations:%d]\n", resp.SessionID, resp.Iterations)
	} else {
		fmt.Fprintf(out, "\n[session:%d used_tools:false]\n", resp.SessionID)
	}
	return nil
}

func runTool(ctx context.Context, out io.Writer, cmd *cobra.Command, name string, input string, sessionID uint) error {
	raw := json.RawMessage(input)
	if !json.Valid(raw) {
		return fmt.Errorf("invalid input JSON")
	}
	database, err := db.Open()
	if err != nil {
		return err
	}
	runner := agent.NewRunner(database, tools.DefaultRegistry())
	allowShell, _ := cmd.Flags().GetBool("allow-shell")
	allowShellDangerous, _ := cmd.Flags().GetBool("allow-shell-dangerous")
	workspace, _ := cmd.Flags().GetString("workspace")
	resp, err := runner.Run(ctx, agent.RunRequest{
		SessionID:           sessionID,
		ToolCalls:           []agent.ToolCall{{Name: name, Input: raw}},
		Workspace:           workspace,
		AllowShell:          allowShell,
		AllowShellDangerous: allowShellDangerous,
	})
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(resp)
}

func runAuditList(out io.Writer, cmd *cobra.Command) error {
	limit, _ := cmd.Flags().GetInt("limit")
	database, err := db.Open()
	if err != nil {
		return err
	}
	events, err := db.ListAuditEvents(database, limit)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(events)
}
func runStatus(out io.Writer) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	payload := map[string]any{"status": "ok", "implementation": "go", "stage": "stage7-telegram", "stack": []string{"cobra", "echo", "gorm", "sqlite"}, "config_dir": must(config.Dir()), "config_path": must(config.Path()), "access_token_path": must(config.AccessTokenPath()), "db_path": must(config.DBPath()), "files_dir": must(config.FilesDir()), "skills_dir": must(config.SkillsDir()), "mcp_path": must(config.MCPPath()), "web_host": settings.WebHost, "web_port": settings.WebPort, "agent_backend": settings.AgentBackend, "model": settings.Model}
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
func runSessionsList(out io.Writer, cmd *cobra.Command) error {
	limit, _ := cmd.Flags().GetInt("limit")
	database, err := db.Open()
	if err != nil {
		return err
	}
	sessions, err := db.ListChatSessions(database, limit)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(sessions)
}
func runSessionsShow(out io.Writer, rawID string) error {
	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || id == 0 {
		return fmt.Errorf("invalid session id %q", rawID)
	}
	database, err := db.Open()
	if err != nil {
		return err
	}
	session, err := db.GetChatSession(database, uint(id))
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(session)
}
func runSessionsDelete(out io.Writer, rawID string) error {
	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || id == 0 {
		return fmt.Errorf("invalid session id %q", rawID)
	}
	database, err := db.Open()
	if err != nil {
		return err
	}
	if err := db.DeleteChatSession(database, uint(id)); err != nil {
		return err
	}
	fmt.Fprintf(out, "deleted session %d\n", id)
	return nil
}

func runSessionsSearch(out io.Writer, cmd *cobra.Command, query string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	database, err := db.Open()
	if err != nil {
		return err
	}
	sessions, err := db.SearchChatSessions(database, query, limit)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(sessions)
}

func runSessionsRename(out io.Writer, rawID string, title string) error {
	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || id == 0 {
		return fmt.Errorf("invalid session id %q", rawID)
	}
	database, err := db.Open()
	if err != nil {
		return err
	}
	session, err := db.RenameChatSession(database, uint(id), title)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(session)
}

func runMemoryList(out io.Writer, cmd *cobra.Command) error {
	limit, _ := cmd.Flags().GetInt("limit")
	asJSON, _ := cmd.Flags().GetBool("json")
	database, err := db.Open()
	if err != nil {
		return err
	}
	items, err := memory.List(database, limit)
	if err != nil {
		return err
	}
	if asJSON {
		return json.NewEncoder(out).Encode(items)
	}
	for _, item := range items {
		fmt.Fprintf(out, "%d\t%s\t%s\n", item.ID, item.Type, item.Content)
	}
	return nil
}

func runMemoryShow(out io.Writer, rawID string) error {
	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || id == 0 {
		return fmt.Errorf("invalid id %q", rawID)
	}
	database, err := db.Open()
	if err != nil {
		return err
	}
	item, err := memory.Get(database, uint(id))
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(item)
}

func runMemoryAdd(out io.Writer, cmd *cobra.Command, itemType string, content string) error {
	metadata, _ := cmd.Flags().GetString("metadata")
	asJSON, _ := cmd.Flags().GetBool("json")
	database, err := db.Open()
	if err != nil {
		return err
	}
	item, err := memory.Add(database, itemType, content, metadata)
	if err != nil {
		return err
	}
	if asJSON {
		return json.NewEncoder(out).Encode(item)
	}
	fmt.Fprintf(out, "added memory %d\n", item.ID)
	return nil
}

func runMemoryDelete(out io.Writer, rawID string) error {
	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || id == 0 {
		return fmt.Errorf("invalid id %q", rawID)
	}
	database, err := db.Open()
	if err != nil {
		return err
	}
	if err := memory.Delete(database, uint(id)); err != nil {
		return err
	}
	fmt.Fprintf(out, "deleted memory %d\n", id)
	return nil
}

func runMemorySearch(out io.Writer, cmd *cobra.Command, query string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	asJSON, _ := cmd.Flags().GetBool("json")
	database, err := db.Open()
	if err != nil {
		return err
	}
	items, err := memory.Search(database, query, limit)
	if err != nil {
		return err
	}
	if asJSON {
		return json.NewEncoder(out).Encode(items)
	}
	for _, item := range items {
		fmt.Fprintf(out, "%d\t%s\t%s\n", item.ID, item.Type, item.Content)
	}
	return nil
}

func runFileStoreList(out io.Writer, cmd *cobra.Command) error {
	limit, _ := cmd.Flags().GetInt("limit")
	asJSON, _ := cmd.Flags().GetBool("json")
	database, err := db.Open()
	if err != nil {
		return err
	}
	items, err := filestore.List(database, limit)
	if err != nil {
		return err
	}
	if asJSON {
		return json.NewEncoder(out).Encode(items)
	}
	for _, item := range items {
		fmt.Fprintf(out, "%d\t%s\t%s\n", item.ID, item.Sha256, item.Name)
	}
	return nil
}

func runFileStoreAdd(out io.Writer, cmd *cobra.Command, path string) error {
	metadata, _ := cmd.Flags().GetString("metadata")
	asJSON, _ := cmd.Flags().GetBool("json")
	database, err := db.Open()
	if err != nil {
		return err
	}
	item, err := filestore.AddFromPath(database, path, metadata)
	if err != nil {
		return err
	}
	if asJSON {
		return json.NewEncoder(out).Encode(item)
	}
	fmt.Fprintf(out, "added file %d\n", item.ID)
	return nil
}

func runFileStoreShow(out io.Writer, rawID string) error {
	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || id == 0 {
		return fmt.Errorf("invalid id %q", rawID)
	}
	database, err := db.Open()
	if err != nil {
		return err
	}
	item, err := filestore.Get(database, uint(id))
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(item)
}

func runFileStoreDelete(out io.Writer, rawID string) error {
	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || id == 0 {
		return fmt.Errorf("invalid id %q", rawID)
	}
	database, err := db.Open()
	if err != nil {
		return err
	}
	if err := filestore.Delete(database, uint(id)); err != nil {
		return err
	}
	fmt.Fprintf(out, "deleted file %d\n", id)
	return nil
}

func runFileStoreSearch(out io.Writer, cmd *cobra.Command, query string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	asJSON, _ := cmd.Flags().GetBool("json")
	database, err := db.Open()
	if err != nil {
		return err
	}
	items, err := filestore.Search(database, query, limit)
	if err != nil {
		return err
	}
	if asJSON {
		return json.NewEncoder(out).Encode(items)
	}
	for _, item := range items {
		fmt.Fprintf(out, "%d\t%s\t%s\n", item.ID, item.Sha256, item.Name)
	}
	return nil
}

func runSearch(out io.Writer, cmd *cobra.Command, query string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	asJSON, _ := cmd.Flags().GetBool("json")
	database, err := db.Open()
	if err != nil {
		return err
	}
	resp, err := search.Run(database, query, limit)
	if err != nil {
		return err
	}
	if asJSON {
		return json.NewEncoder(out).Encode(resp)
	}
	for _, r := range resp.Results {
		fmt.Fprintf(out, "%s\t%d\t%s\t%s\n", r.Type, r.ID, r.Title, r.Snippet)
	}
	return nil
}

func runSkillsList(out io.Writer) error {
	report, err := skills.LoadAll()
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(map[string]any{
		"root":   report.Root,
		"skills": report.Skills,
		"errors": report.Errors,
	})
}

func runSkillsReload(out io.Writer) error {
	return runSkillsList(out)
}

func runSkillsShow(out io.Writer, name string) error {
	s, err := skills.LoadByName(name)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(s)
}

func runSkillsValidate(out io.Writer, path string) error {
	s, err := skills.LoadFromFile(path)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(map[string]any{"valid": true, "skill": s})
}

func runSkillsInstall(out io.Writer, cmd *cobra.Command, path string) error {
	force, _ := cmd.Flags().GetBool("force")
	s, err := skills.InstallFromDir(path, skills.InstallOptions{Force: force})
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(map[string]any{"installed": true, "skill": s})
}

func runSkillsUninstall(out io.Writer, cmd *cobra.Command, name string) error {
	yes, _ := cmd.Flags().GetBool("yes")
	if err := skills.Uninstall(name, skills.UninstallOptions{Yes: yes}); err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(map[string]any{"uninstalled": true, "name": name})
}

var mcpManager = mcp.NewManager()

func runMCPList(out io.Writer) error {
	cfg, err := mcp.LoadConfig()
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(map[string]any{"servers": cfg.Servers, "names": mcp.ListServerNames(cfg)})
}

func runMCPShow(out io.Writer, name string) error {
	cfg, err := mcp.LoadConfig()
	if err != nil {
		return err
	}
	if err := mcp.ValidateName(name); err != nil {
		return err
	}
	server, ok := cfg.Servers[name]
	if !ok {
		return fmt.Errorf("unknown server %q", name)
	}
	return json.NewEncoder(out).Encode(map[string]any{"name": name, "config": server})
}

func runMCPAdd(out io.Writer, cmd *cobra.Command, name string) error {
	if err := mcp.ValidateName(name); err != nil {
		return err
	}
	command, _ := cmd.Flags().GetString("command")
	args, _ := cmd.Flags().GetStringArray("arg")
	cfg, err := mcp.LoadConfig()
	if err != nil {
		return err
	}
	cfg.Servers[name] = mcp.ServerConfig{Command: command, Args: args, Env: map[string]string{}}
	if err := mcp.SaveConfig(cfg); err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(map[string]any{"saved": true, "name": name})
}

func runMCPRemove(out io.Writer, name string) error {
	if err := mcp.ValidateName(name); err != nil {
		return err
	}
	cfg, err := mcp.LoadConfig()
	if err != nil {
		return err
	}
	delete(cfg.Servers, name)
	if err := mcp.SaveConfig(cfg); err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(map[string]any{"removed": true, "name": name})
}

func runMCPStart(out io.Writer, name string) error {
	cfg, err := mcp.LoadConfig()
	if err != nil {
		return err
	}
	server, ok := cfg.Servers[name]
	if !ok {
		return fmt.Errorf("unknown server %q", name)
	}
	st, err := mcpManager.Start(context.Background(), name, server)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(st)
}

func runMCPStop(out io.Writer, name string) error {
	st, err := mcpManager.Stop(context.Background(), name)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(st)
}

func runMCPStatus(out io.Writer) error {
	return json.NewEncoder(out).Encode(mcpManager.Status())
}

func runMCPPresets(out io.Writer) error {
	return json.NewEncoder(out).Encode(presets.ListMCPPresets())
}

func runMCPInstallPreset(out io.Writer, cmd *cobra.Command, presetName string) error {
	workspace, _ := cmd.Flags().GetString("workspace")
	cfg, err := mcp.LoadConfig()
	if err != nil {
		return err
	}
	serverCfg, err := presets.BuildMCPServerConfig(presetName, workspace)
	if err != nil {
		return err
	}
	if err := mcp.ValidateName(presetName); err != nil {
		return err
	}
	cfg.Servers[presetName] = serverCfg
	if err := mcp.SaveConfig(cfg); err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(map[string]any{"installed": true, "name": presetName, "config": serverCfg})
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
func toLLMMessages(messages []db.ChatMessage) []llm.Message {
	out := make([]llm.Message, 0, len(messages))
	for _, message := range messages {
		if message.Role == "" || message.Content == "" {
			continue
		}
		out = append(out, llm.Message{Role: message.Role, Content: message.Content})
	}
	return out
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
