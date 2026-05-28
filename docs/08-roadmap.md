# Roadmap: Go core stages 5–8

Этот документ заменяет корневой `todo.md` как основной roadmap для продолжения Go-порта Paw.

Статус сверялся по текущему Go-коду, а не только по старому TODO. На момент проверки CLI уже собирает команды `sessions`, `memory`, `file-store`, `search`, `skills`, `mcp`, `channels`, `tools`, `run-tool`, `audit`; HTTP server уже регистрирует маршруты для sessions, memory, files, search, skills, mcp и channels.

## Легенда

- `[done]` — реализовано в Go core.
- `[partial]` — базовая реализация есть, но есть открытые ограничения или нужны тесты/полировка.
- `[todo]` — пока не реализовано или сознательно отложено.
- `[deferred]` — не приоритет для ближайшего этапа.

## Global rules

- Не трогать существующие Python `src/` и `ee/`, кроме использования как reference.
- Сохранять Go stack: Cobra для CLI, Echo для HTTP/WebSocket API, GORM + SQLite для local persistence.
- Не ломать существующие CLI commands и API routes.
- Делать small focused PRs.
- Добавлять docs и tests где практично.
- Тесты не должны требовать реальные API keys.
- Перед сдачей запускать:

```bash
go test ./...
```

## Current implementation snapshot

### Уже есть

- `[done]` CLI root: `paw serve|chat|agent|status|doctor|config|auth|db|sessions|memory|file-store|search|skills|mcp|channels|tools|run-tool|audit`.
- `[done]` API: health/status/settings/chat/sessions/tools/agent/audit + memory/files/search + skills + mcp + channels + `/ws`.
- `[done]` LLM providers: Ollama, OpenAI-compatible, OpenAI, Anthropic.
- `[done]` Agent: multi-iteration tool loop, workspace sandboxing/policy, tool audit, WebSocket progress events.
- `[done]` Storage: SQLite database at `~/.pocketpaw/paw.db`, file store at `~/.pocketpaw/files`.
- `[done]` Skills: YAML format, load/list/show/validate/reload, install/uninstall, injection into agent with `paw agent --skill`.
- `[done]` MCP: `~/.pocketpaw/mcp.json`, list/show/add/remove/start/stop/status, built-in presets including filesystem.
- `[partial]` Channels: core manager and Telegram adapter exist; Discord/Slack intentionally not implemented yet.
- `[done]` Channel audit: Telegram writes `channel.message.received`, `channel.message.sent`, `channel.error` audit events.

### Главные оставшиеся gaps

- `[todo]` OAuth2/scoped auth, API keys, identity flows compatible with Python API.
- `[todo]` Integrations/connectors/kits.
- `[todo]` Metrics/analytics/traces/events/alerts.
- `[todo]` Reminders/proactive/daemon/scheduler.
- `[todo]` Remote tunnel / remote control.
- `[todo]` Vector DB / embeddings retrieval.
- `[todo]` Browser automation.
- `[deferred]` Soul/deep work/mission control/A2A/fabric-level orchestration.
- `[deferred]` EE implementation beyond docs/boundaries.

## Stage 5 — Memory / Sessions / File store / Search

### Goal

Превратить session storage в usable memory, file store и local search layer.

### Status

`[done]` Stage 5 в основном реализован.

### 5.1 Memory CLI

Required commands:

```bash
paw memory list
paw memory show <id>
paw memory add <type> <content>
paw memory delete <id>
paw memory search <query>
```

Implemented:

- `[done]` `paw memory list`, alias `ls`.
- `[done]` `paw memory show <id>`.
- `[done]` `paw memory add <type> <content>`.
- `[done]` `paw memory delete <id>`, alias `rm`.
- `[done]` `paw memory search <query>`.
- `[done]` JSON output flags for list/add/search.

### 5.2 Memory API

Implemented routes:

```text
GET    /api/v1/memory
POST   /api/v1/memory
GET    /api/v1/memory/:id
DELETE /api/v1/memory/:id
GET    /api/v1/memory/search?q=...
```

Status:

- `[done]` Routes exist and are behind the shared access token middleware.
- `[done]` CRUD and search are present.
- `[partial]` Error shape should be reviewed for consistency across all endpoints.

### 5.3 Session improvements

Implemented CLI:

```bash
paw sessions list
paw sessions show <id>
paw sessions search <query>
paw sessions rename <id> <title>
paw sessions delete <id>
```

Implemented API:

```text
GET    /api/v1/sessions
GET    /api/v1/sessions/:id
DELETE /api/v1/sessions/:id
GET    /api/v1/sessions/search?q=...
PATCH  /api/v1/sessions/:id
```

Status:

- `[done]` Search by title/messages.
- `[done]` Rename through CLI/API.
- `[partial]` Need integration tests for search result shape and rename edge cases.

### 5.4 File store

Implemented CLI:

```bash
paw file-store list
paw file-store add <path>
paw file-store show <id>
paw file-store delete <id>
paw file-store search <query>
```

Implemented API:

```text
GET    /api/v1/files
POST   /api/v1/files
GET    /api/v1/files/:id
DELETE /api/v1/files/:id
GET    /api/v1/files/search?q=...
```

Status:

- `[done]` Imported files are stored under `~/.pocketpaw/files`.
- `[done]` File metadata is persisted in SQLite.
- `[done]` File records can be listed, shown, searched and deleted.
- `[partial]` Confirm test coverage for SHA-256, path traversal denial, and delete cleanup.

### 5.5 Global search

Implemented CLI/API:

```bash
paw search <query>
```

```text
GET /api/v1/search?q=...
```

Status:

- `[done]` Searches across memory, sessions, messages, and file records.
- `[partial]` Current implementation is basic local SQL-style search; embeddings/vector DB are intentionally later.

### Stage 5 remaining work

- `[todo]` Add/verify tests for memory CRUD/search.
- `[todo]` Add/verify tests for session rename/search.
- `[todo]` Add/verify tests for file store add/list/delete and path traversal denial.
- `[todo]` Add/verify tests for global search result shape.

## Stage 6 — Skills / MCP

### Goal

Add reusable skills and local MCP server management.

### Status

`[done]` Stage 6 baseline is implemented.

### 6.1 Skills format

Status:

- `[done]` Local skill directory is `~/.pocketpaw/skills/<skill-name>/skill.yaml`.
- `[done]` Skill validation exists.
- `[done]` Skills are not executed automatically; they are selected explicitly.

Minimum expected fields remain:

```text
name
description
version
prompts
commands
```

### 6.2 Skills loader

Implemented CLI:

```bash
paw skills list
paw skills show <name>
paw skills validate <path>
paw skills reload
```

Implemented API:

```text
GET  /api/v1/skills
GET  /api/v1/skills/:name
POST /api/v1/skills/reload
```

Status:

- `[done]` Skills can be discovered from `~/.pocketpaw/skills`.
- `[done]` Bad skill files return validation errors instead of crashing the app.

### 6.3 Skills installer

Implemented CLI:

```bash
paw skills install <path>
paw skills uninstall <name>
```

Status:

- `[done]` Local directory install.
- `[done]` Uninstall with `--yes` confirmation bypass.
- `[done]` Existing skills require `--force` to overwrite.
- `[deferred]` Git URL install is not required for this stage.

### 6.4 Skill integration with agent

Implemented CLI:

```bash
paw agent --skill go-reviewer "Review this project"
```

Status:

- `[done]` Skill system prompt can be injected into agent context.
- `[done]` Missing skill returns an error.
- `[partial]` Audit visibility for skill usage should be checked and made explicit if missing.

### 6.5 MCP manager

Implemented config path:

```text
~/.pocketpaw/mcp.json
```

Implemented CLI:

```bash
paw mcp list
paw mcp show <name>
paw mcp add <name> --command <cmd> --arg <arg>...
paw mcp remove <name>
paw mcp start <name>
paw mcp stop <name>
paw mcp status
```

Implemented API:

```text
GET    /api/v1/mcp
GET    /api/v1/mcp/:name
POST   /api/v1/mcp
DELETE /api/v1/mcp/:name
POST   /api/v1/mcp/:name/start
POST   /api/v1/mcp/:name/stop
GET    /api/v1/mcp/status
```

Status:

- `[done]` MCP servers can be added/removed from config.
- `[done]` Process lifecycle start/stop/status exists.
- `[partial]` Full MCP tool protocol integration can stay incremental.

### 6.6 MCP presets

Implemented CLI:

```bash
paw mcp presets
paw mcp install-preset filesystem --workspace .
```

Status:

- `[done]` Presets can be listed.
- `[done]` Preset install writes an MCP server config.
- `[partial]` Current known preset baseline is filesystem; add git/github/sqlite later only if needed.

### Stage 6 remaining work

- `[todo]` Add/verify tests for skill YAML parsing and validation.
- `[todo]` Add/verify tests for skill install/uninstall paths.
- `[todo]` Add/verify tests for MCP config load/save.
- `[todo]` Add/verify tests for MCP preset generation.

## Stage 7 — Channels

### Goal

Add external communication channels as adapters. Telegram first; Discord/Slack later only if useful.

### Status

`[partial]` Channel core and Telegram are implemented. Discord/Slack are deferred.

### 7.1 Channel core

Implemented interface:

```go
type Channel interface {
    Name() string
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Status() ChannelStatus
}
```

Implemented CLI/API:

```bash
paw channels list
paw channels status
paw channels start <name>
paw channels stop <name>
```

```text
GET  /api/v1/channels
GET  /api/v1/channels/status
POST /api/v1/channels/:name/start
POST /api/v1/channels/:name/stop
```

Status:

- `[done]` Channel manager can register adapters.
- `[done]` Start/stop/status works for registered adapters.
- `[done]` CLI/API expose consistent channel status surface.

### 7.2 Telegram adapter

Config keys:

```text
telegram_bot_token
allowed_user_id
```

Status:

- `[done]` Adapter starts only when `telegram_bot_token` is configured.
- `[done]` Optional `allowed_user_id` authorization check exists.
- `[done]` Incoming messages are routed through the Go agent runner.
- `[done]` Replies are sent back to Telegram.
- `[done]` Status includes running/error state.
- `[partial]` Needs fake-adapter or mocked Telegram tests.

### 7.3 Discord adapter

Status: `[deferred]`

Not implemented. Keep skipped unless Telegram is stable and there is an actual user need.

Potential config keys:

```text
discord_bot_token
discord_allowed_guild_id
discord_allowed_channel_id
```

### 7.4 Slack adapter

Status: `[deferred]`

Not implemented. Keep skipped unless Telegram/Discord are stable and there is an actual user need.

Potential config keys:

```text
slack_bot_token
slack_app_token
slack_allowed_channel_id
```

### 7.5 Channel persistence / audit

Status:

- `[done]` Telegram emits `channel.message.received`.
- `[done]` Telegram emits `channel.message.sent`.
- `[done]` Telegram emits `channel.error`.
- `[done]` Tokens are not included in channel audit payloads.
- `[partial]` Generalize/verify audit behavior for future non-Telegram adapters.

### Stage 7 remaining work

- `[todo]` Add tests for channel manager registration and lifecycle.
- `[todo]` Add tests for Telegram allowed-user checks.
- `[todo]` Add tests for channel audit event creation.
- `[deferred]` Discord adapter.
- `[deferred]` Slack adapter.

## Stage 8 — Enterprise / EE

### Goal

Only start EE after the core is stable. EE remains documentation/design first.

### Status

`[deferred]` Do not implement EE runtime features yet unless explicitly requested.

EE areas:

```text
ee/cloud
ee/fleet
ee/pawprint
ee/fabric
```

Rules:

- Keep open-core boundaries clear.
- Do not mix proprietary EE code into core packages unless explicitly designed as interfaces.
- Core must remain useful without EE.
- EE should depend on core interfaces, not the other way around.

### 8.1 EE architecture boundary

Status: `[todo]` docs/interface boundary only.

Acceptance criteria:

- Core builds without EE-specific services.
- EE packages do not break normal `go test ./...`.
- Any build tags are documented.

### 8.2 `ee/cloud`

Purpose:

```text
Cloud sync, hosted account integration, remote storage, remote settings.
```

Status: `[deferred]`

### 8.3 `ee/fleet`

Purpose:

```text
Manage multiple Paw nodes/agents.
```

Status: `[deferred]`

### 8.4 `ee/pawprint`

Purpose:

```text
Portable environment/project snapshots.
```

Status: `[deferred]`

### 8.5 `ee/fabric`

Purpose:

```text
Higher-level orchestration fabric for agents, skills, channels, memory, and fleet.
```

Status: `[deferred]`

## Cross-stage acceptance criteria

Before considering stages 5–8 ready:

```bash
go test ./...
```

Manual smoke checklist:

```bash
paw status
paw doctor
paw memory add fact "The project is called Paw."
paw memory search Paw
paw sessions list
paw sessions search Paw
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

## Suggested next PRs

1. `tests-stage5-memory-files-search`
2. `tests-stage6-skills-mcp`
3. `tests-stage7-channels-telegram`
4. `docs-ee-boundaries`
5. `oauth-auth-design`
6. `minimal-ui-required-api-audit`
7. `vector-search-design`

## Definition of done

Stages 5–8 are complete when:

- `[done]` Memory CRUD/search works through CLI and API.
- `[done]` Sessions support list/show/search/rename/delete.
- `[done]` File store supports add/list/show/search/delete.
- `[done]` Global search covers memory, sessions, messages, and file records.
- `[done]` Skills can be loaded, validated, installed, listed, shown, and used by the agent.
- `[done]` MCP configs can be managed and process lifecycle can be started/stopped/statused.
- `[done]` MCP presets can be listed and installed.
- `[done]` Channel manager exists and Telegram adapter works.
- `[deferred]` Discord and Slack adapters are implemented only after Telegram and only if needed.
- `[done]` Channel status is available through CLI and API.
- `[done]` Telegram channel activity is audited.
- `[todo]` EE directories contain clear docs and boundaries after core stability.
- `[todo]` `go test ./...` passes in CI/local verification.
