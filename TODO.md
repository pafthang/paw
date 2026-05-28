# TODO — Codex Roadmap for Paw Go Core Stages 5–8

This document is a task specification for Codex to continue the Paw Go port after the first four core stages.

The current Go core already contains the foundation for:

- CLI / config / server / health
- API compatibility with token auth and WebSocket route
- LLM clients for Ollama and OpenAI-compatible providers
- sessions, tools, shell/file tools, audit log, and initial agent loop

Stages 5–8 should build on top of that foundation without rewriting the existing implementation.

## Global rules

- Keep the existing Python `src/` and `ee/` directories untouched unless they are used only as references.
- Keep the Go stack:
  - Cobra for CLI
  - Echo for HTTP/WebSocket API
  - GORM + SQLite for local persistence
- Preserve existing CLI commands and API routes.
- Prefer small, focused PRs.
- Every stage should include docs and tests where practical.
- Do not require real API keys in tests.
- Run before submitting:

```bash
go test ./...
```

---

# Stage 5 — Memory / Sessions

## Goal

Turn the existing session storage into a usable memory and file store layer.

Stage 5 should provide:

```text
paw memory
paw sessions
file_store
search
delete
```

Existing session support should be preserved and expanded.

## Current baseline

Already present in the Go core:

- `ChatSession`
- `ChatMessage`
- `MemoryItem`
- `paw sessions list/show/delete`
- SQLite database at `~/.pocketpaw/paw.db`
- session-aware `paw chat`
- session-aware `paw agent`
- session-aware HTTP and WebSocket paths

## Required packages / areas

Suggested packages:

```text
internal/memory
internal/filestore
internal/search
```

Existing packages to inspect:

```text
internal/db
internal/cli
internal/server
internal/agent
internal/tools
```

## 5.1 Memory CLI

Add a first-class `paw memory` command group.

Required commands:

```bash
paw memory list
paw memory show <id>
paw memory add <type> <content>
paw memory delete <id>
paw memory search <query>
```

Optional useful aliases:

```bash
paw memory ls
paw memory rm <id>
```

Expected behavior:

- `paw memory add` stores into `MemoryItem`.
- `paw memory list` returns recent memory items.
- `paw memory search` searches content and metadata.
- `paw memory delete` removes a memory item.

JSON output should be available where useful:

```bash
paw memory list --json
paw memory search "project name" --json
```

Acceptance criteria:

- CLI commands compile and work.
- Memory entries persist in SQLite.
- Memory search returns relevant entries.
- Deleting a memory item removes it from future list/search results.

## 5.2 Memory API

Add HTTP API routes:

```text
GET    /api/v1/memory
POST   /api/v1/memory
GET    /api/v1/memory/:id
DELETE /api/v1/memory/:id
GET    /api/v1/memory/search?q=...
```

Protected by existing access token middleware.

Expected JSON shapes:

```json
{
  "type": "fact",
  "content": "The project is called Paw.",
  "metadata": {
    "source": "manual"
  }
}
```

Acceptance criteria:

- Routes require auth token.
- CRUD works.
- Search works.
- Errors are JSON and actionable.

## 5.3 Session improvements

Extend `paw sessions` with search and rename.

Required commands:

```bash
paw sessions search <query>
paw sessions rename <id> <title>
paw sessions delete <id>
```

Required API additions:

```text
GET   /api/v1/sessions/search?q=...
PATCH /api/v1/sessions/:id
```

Patch payload:

```json
{
  "title": "New title"
}
```

Acceptance criteria:

- Session title can be changed.
- Session search searches title and messages.
- Search results include enough context to select a session.

## 5.4 File store

Add a local file store for artifacts and durable project files managed by Paw.

Default storage root:

```text
~/.pocketpaw/files
```

Suggested model:

```go
type FileRecord struct {
    ID        uint
    Path      string
    Name      string
    MimeType  string
    SizeBytes int64
    Sha256    string
    Metadata  string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

Required CLI:

```bash
paw file-store list
paw file-store add <path>
paw file-store show <id>
paw file-store delete <id>
paw file-store search <query>
```

Required API:

```text
GET    /api/v1/files
POST   /api/v1/files
GET    /api/v1/files/:id
DELETE /api/v1/files/:id
GET    /api/v1/files/search?q=...
```

Implementation notes:

- Copy added files into `~/.pocketpaw/files`.
- Store metadata in SQLite.
- Compute SHA-256.
- Prevent path traversal.
- Do not expose arbitrary host files through API.

Acceptance criteria:

- Files can be imported into file store.
- Stored files survive process restart.
- Files can be listed, searched, and deleted.
- Deleting removes DB record and stored file.

## 5.5 Search

Implement basic local search first.

Minimum search scope:

```text
memory items
session titles
chat messages
file records
```

Required CLI:

```bash
paw search <query>
```

Required API:

```text
GET /api/v1/search?q=...
```

Expected response shape:

```json
{
  "query": "paw",
  "results": [
    {
      "type": "memory",
      "id": 1,
      "title": "fact",
      "snippet": "The project is called Paw."
    }
  ]
}
```

Implementation can start with SQL `LIKE` queries. Do not require embeddings yet.

Acceptance criteria:

- `paw search` searches across memory, sessions, messages, and file records.
- API search works with auth.
- Search result types are clear.

## Stage 5 tests

Add tests for:

- memory CRUD
- memory search
- session rename/search
- file store add/list/delete
- path traversal denial
- global search result shape

---

# Stage 6 — Skills / MCP

## Goal

Add a skills system and MCP manager so Paw can load reusable capabilities and external tool servers.

Stage 6 should provide:

```text
skills loader
skills installer
MCP manager
MCP presets
```

## Required packages / areas

Suggested packages:

```text
internal/skills
internal/mcp
internal/presets
```

Existing packages to integrate with:

```text
internal/tools
internal/agent
internal/config
internal/cli
internal/server
internal/db
```

## 6.1 Skills format

Define a local skill format.

Suggested directory:

```text
~/.pocketpaw/skills/<skill-name>/skill.yaml
```

Suggested `skill.yaml`:

```yaml
name: go-reviewer
description: Helps review Go code and run tests.
version: 0.1.0
prompts:
  system: |
    You are a Go code reviewer.
commands:
  - name: test
    description: Run Go tests
    tool: shell.run
    input:
      command: go test ./...
      timeout_seconds: 120
```

Minimum fields:

```text
name
description
version
prompts
commands
```

Acceptance criteria:

- Invalid skill files return clear errors.
- Skills can be loaded from local directory.
- Skills are not executed automatically without explicit user/agent choice.

## 6.2 Skills loader

Required CLI:

```bash
paw skills list
paw skills show <name>
paw skills validate <path>
paw skills reload
```

Required API:

```text
GET /api/v1/skills
GET /api/v1/skills/:name
POST /api/v1/skills/reload
```

Acceptance criteria:

- Skills can be discovered from `~/.pocketpaw/skills`.
- Bad skills do not crash the app.
- Loader returns useful validation errors.

## 6.3 Skills installer

Support installing skills from:

```text
local directory
local archive if practical
Git URL later, optional for this stage
```

Required CLI:

```bash
paw skills install <path-or-url>
paw skills uninstall <name>
```

Minimum behavior:

- Local directory install copies into `~/.pocketpaw/skills/<name>`.
- Uninstall removes skill directory after confirmation or `--yes`.
- Validate before install.

Acceptance criteria:

- Local skill install works.
- Uninstall works.
- Existing skills are not overwritten unless `--force` is passed.

## 6.4 Skill integration with agent

Add ability to run agent with skill context.

CLI:

```bash
paw agent --skill go-reviewer "Review this project"
```

API/WS:

```json
{
  "skill": "go-reviewer",
  "prompt": "Review this project"
}
```

Expected behavior:

- Skill system prompt is merged into agent system prompt.
- Skill commands can be exposed as recommended tool calls or helper presets.
- Audit should record skill name when used.

Acceptance criteria:

- `paw agent --skill <name>` changes agent behavior by injecting skill prompt.
- Missing skill returns clear error.
- Skill usage is visible in response or audit.

## 6.5 MCP manager

Add a local MCP server registry/manager.

Suggested config file:

```text
~/.pocketpaw/mcp.json
```

Suggested schema:

```json
{
  "servers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/project"],
      "env": {}
    }
  }
}
```

Required CLI:

```bash
paw mcp list
paw mcp show <name>
paw mcp add <name> --command <cmd> --arg <arg>...
paw mcp remove <name>
paw mcp start <name>
paw mcp stop <name>
paw mcp status
```

Required API:

```text
GET    /api/v1/mcp
GET    /api/v1/mcp/:name
POST   /api/v1/mcp
DELETE /api/v1/mcp/:name
POST   /api/v1/mcp/:name/start
POST   /api/v1/mcp/:name/stop
GET    /api/v1/mcp/status
```

Initial implementation can manage config and process lifecycle. Full MCP tool protocol integration can be incremental.

Acceptance criteria:

- MCP servers can be added/removed from config.
- `start/stop/status` works for local process-based servers.
- Errors are clear when command is missing or process exits.

## 6.6 MCP presets

Add built-in presets for common MCP servers.

Suggested presets:

```text
filesystem
git
github
sqlite
browser optional later
```

Required CLI:

```bash
paw mcp presets
paw mcp install-preset filesystem --workspace .
```

Acceptance criteria:

- Presets are listed with descriptions.
- Installing a preset adds an MCP server config.
- Presets do not require secrets unless clearly documented.

## Stage 6 tests

Add tests for:

- skill YAML parsing
- skill validation
- skill install/uninstall paths
- MCP config load/save
- MCP preset generation

---

# Stage 7 — Channels

## Goal

Add external communication channels as adapters.

Start with Telegram, then Discord and Slack.

All channels should implement:

```go
type Channel interface {
    Name() string
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Status() ChannelStatus
}
```

## Required packages / areas

Suggested packages:

```text
internal/channels
internal/channels/telegram
internal/channels/discord
internal/channels/slack
```

Integrate with:

```text
internal/agent
internal/config
internal/db
internal/cli
internal/server
```

## 7.1 Channel core

Define:

```go
type Channel interface {
    Name() string
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Status() ChannelStatus
}

type ChannelStatus struct {
    Name      string
    Running   bool
    LastError string
    StartedAt time.Time
}
```

Add manager:

```go
type Manager struct {
    channels map[string]Channel
}
```

Required CLI:

```bash
paw channels list
paw channels status
paw channels start <name>
paw channels stop <name>
```

Required API:

```text
GET  /api/v1/channels
GET  /api/v1/channels/status
POST /api/v1/channels/:name/start
POST /api/v1/channels/:name/stop
```

Acceptance criteria:

- Channel manager can register adapters.
- Start/stop/status works for adapters.
- API and CLI report consistent status.

## 7.2 Telegram adapter first

Config keys already include or should include:

```text
telegram_bot_token
allowed_user_id
```

Expected behavior:

- Telegram bot starts when configured.
- Only `allowed_user_id` can interact if set.
- Incoming Telegram messages are routed to `agent.Chat` or `chat` path.
- Replies are sent back to Telegram.
- Errors are logged and visible in channel status.

Required CLI:

```bash
paw channels start telegram
paw channels stop telegram
paw channels status
```

Optional helper:

```bash
paw telegram status
```

Acceptance criteria:

- Missing token returns clear error.
- Unauthorized user is ignored or receives a denial message.
- Authorized user can send a prompt and receive a reply.
- Channel status reports running/error state.

## 7.3 Discord adapter

Add after Telegram is stable.

Config keys:

```text
discord_bot_token
discord_allowed_guild_id
discord_allowed_channel_id
```

Acceptance criteria:

- Missing token returns clear error.
- Bot responds only in allowed guild/channel when configured.
- Messages route to agent/chat path.

## 7.4 Slack adapter

Add after Discord.

Config keys:

```text
slack_bot_token
slack_app_token optional
slack_allowed_channel_id
```

Acceptance criteria:

- Missing token returns clear error.
- Bot responds only in allowed channel when configured.
- Messages route to agent/chat path.

## 7.5 Channel persistence / audit

Add audit events for channel activity:

```text
channel.message.received
channel.message.sent
channel.error
```

Acceptance criteria:

- Channel activity is visible in `paw audit list`.
- Sensitive tokens are never stored in audit output.

## Stage 7 tests

Use fake adapters where external services would be needed.

Add tests for:

- channel manager registration
- start/stop/status lifecycle
- auth/allowed-user checks for Telegram logic if isolated
- audit event creation

---

# Stage 8 — Enterprise / EE

## Goal

Only start EE after the core is stable.

EE areas:

```text
ee/cloud
ee/fleet
ee/pawprint
ee/fabric
```

Important: do not begin implementing EE until stages 1–7 are stable enough.

## Stage 8 rules

- Keep open-core boundaries clear.
- Do not mix proprietary EE code into core packages unless explicitly designed as interfaces.
- Core must remain useful without EE.
- EE should depend on core interfaces, not the other way around.

## 8.1 EE architecture boundary

Define interfaces in core where needed, for example:

```text
internal/core interfaces, or pkg/paw if public API is desired later
```

EE should live under:

```text
ee/cloud
ee/fleet
ee/pawprint
ee/fabric
```

Acceptance criteria:

- Core builds without EE-specific services.
- EE packages do not break normal `go test ./...`.
- Any build tags are documented.

## 8.2 ee/cloud

Purpose:

```text
Cloud sync, hosted account integration, remote storage, remote settings.
```

Possible future tasks:

- account auth
- sync sessions/memory/file_store
- cloud backup/restore
- remote model provider config

Initial Codex task should only create interfaces and docs unless core is stable.

## 8.3 ee/fleet

Purpose:

```text
Manage multiple Paw nodes/agents.
```

Possible future tasks:

- node registration
- node status
- remote command dispatch
- fleet audit log
- distributed tool execution policy

Initial Codex task should only define design docs and interfaces.

## 8.4 ee/pawprint

Purpose:

```text
Portable environment/project snapshots.
```

Possible future tasks:

- export/import project context
- dependency metadata
- model/tool/skill bundle metadata
- reproducible Paw workspace snapshot

Initial Codex task should only define format draft and docs.

## 8.5 ee/fabric

Purpose:

```text
Higher-level orchestration fabric for agents, skills, channels, memory, and fleet.
```

Possible future tasks:

- workflow graph
- multi-agent coordination
- scheduled jobs
- shared policy engine

Initial Codex task should only define architecture after core stabilizes.

## Stage 8 acceptance criteria

For now, Stage 8 is documentation/design only unless the maintainer explicitly requests implementation.

Required docs:

```text
ee/README.md
ee/cloud/README.md
ee/fleet/README.md
ee/pawprint/README.md
ee/fabric/README.md
```

Do not add cloud credentials, proprietary endpoints, or hard-coded vendor assumptions.

---

# Cross-stage acceptance criteria

Before considering stages 5–8 ready:

```bash
go test ./...
```

Must pass without real external tokens.

Manual smoke checklist:

```bash
paw status
paw doctor
paw memory add fact "The project is called Paw."
paw memory search Paw
paw sessions list
paw search Paw
paw skills list
paw mcp presets
paw channels status
```

HTTP smoke checklist:

```bash
TOKEN=$(paw auth token)

curl -s http://127.0.0.1:8888/api/v1/memory \
  -H "Authorization: Bearer $TOKEN"

curl -s "http://127.0.0.1:8888/api/v1/search?q=Paw" \
  -H "Authorization: Bearer $TOKEN"

curl -s http://127.0.0.1:8888/api/v1/skills \
  -H "Authorization: Bearer $TOKEN"

curl -s http://127.0.0.1:8888/api/v1/channels/status \
  -H "Authorization: Bearer $TOKEN"
```

# Suggested PR breakdown

Prefer these PRs:

1. `stage5-memory-crud`
2. `stage5-file-store`
3. `stage5-global-search`
4. `stage6-skills-loader`
5. `stage6-skills-installer`
6. `stage6-mcp-manager`
7. `stage6-mcp-presets`
8. `stage7-channel-core`
9. `stage7-telegram-channel`
10. `stage7-discord-channel`
11. `stage7-slack-channel`
12. `stage8-ee-docs-and-boundaries`

# Definition of done

Stages 5–8 are complete when:

- Memory CRUD/search works through CLI and API.
- Sessions support list/show/search/rename/delete.
- File store supports add/list/show/search/delete.
- Global search covers memory, sessions, messages, and file records.
- Skills can be loaded, validated, installed, listed, shown, and used by the agent.
- MCP configs can be managed and process lifecycle can be started/stopped/statused.
- MCP presets can be listed and installed.
- Channel manager exists and at least Telegram adapter works.
- Discord and Slack adapters are implemented after Telegram.
- Channel status is available through CLI and API.
- Channel activity is audited.
- EE directories contain clear docs and boundaries after core stability.
- `go test ./...` passes.
