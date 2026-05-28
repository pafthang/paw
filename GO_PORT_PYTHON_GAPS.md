# Go Port — Python → Go gap analysis (as of stage7-telegram)

This repo contains a large Python implementation under `src/pocketpaw/`. The Go core under `internal/` currently implements a focused subset (CLI + API compatibility + basic agent/tooling + storage + skills/MCP skeleton + channels skeleton + Telegram).

This document inventories major Python areas that are *not yet* ported to Go, and suggests a pragmatic porting order.

## What Go already covers (high level)

- CLI: `paw serve|chat|agent|status|doctor|config|auth|db|sessions|memory|file-store|search|skills|mcp|channels|tools|run-tool|audit`
- API: health/status/settings/chat/sessions/tools/agent/audit + memory/files/search + skills + mcp + channels + `/ws`
- LLM providers: Ollama, OpenAI-compatible, OpenAI, Anthropic
- Agent: multi-iteration tool loop, workspace sandboxing/policy, audit of tool runs/denials, WS progress events
- Storage: SQLite (`~/.pocketpaw/paw.db`), file store (`~/.pocketpaw/files`)
- Skills: YAML format, load/list/show/validate/reload + install/uninstall + inject into agent
- MCP: `mcp.json` config + start/stop/status + filesystem preset
- Channels: core interface/manager + Telegram adapter baseline

## Python areas not yet ported (grouped)

### API surface beyond Go compatibility subset

Python `src/pocketpaw/api/v1/*` includes many routers the Go server does not implement yet:

- OAuth2 / scoped auth (`oauth2.py`, `api_keys.py`, `identity.py`, `auth.py`)
- Integrations / connectors / kits (`connectors.py`, `kits.py`, `oauth_integrations.py`)
- Operational/telemetry (`metrics.py`, `analytics.py`, `traces.py`, `events.py`, `alerts.py`)
- Reminders / proactive / plan-mode (`reminders.py`, `plan_mode.py`, plus `daemon/*`)
- Remote tunnel / remote control (`remote.py`)
- Soul / deep work / mission control (`soul.py`, `deep_work/*`, `mission_control/*`)

Suggested approach: do not blindly port endpoints; first decide which ones are required for Go milestone or UI clients.

### Vector DB / embeddings

- `vectordb/*` (e.g. Chroma adapter) and related retrieval.

Likely Stage 9+ unless search relevance is insufficient with SQL `LIKE`.

### Browser automation

- `browser/*` in Python implies a larger subsystem (playwright/selenium-like).

Port only if explicitly needed; it impacts sandbox/policy heavily.

### Daemon / background tasks

- `daemon/*` (triggers, proactive executor) and likely scheduled jobs.

In Go, this should be a separate package + opt-in start from CLI, and share the audit/policy system.

### Additional channels

- Slack / Discord adapters exist conceptually in Python and are specified in `todo.md`.

### “Soul protocol”, A2A, mission control

Python contains `soul`, `a2a`, and agent orchestration layers not present in Go yet.

Porting these should come *after* core storage/policy/channels stabilize in Go, otherwise the surface area balloons.

## Recommended porting order (after current state)

1. Finish Stage 7 channels: Discord, Slack + channel audit events and stable status reporting.
2. Stabilize OAuth / auth story: decide whether Go keeps “access_token file” or adopts Python’s OAuth2 scopes for UI clients.
3. Port only the minimum extra API endpoints required by the dashboard/UI (if any).
4. Add daemon/scheduler only if required for reminders/proactive flows.
5. Consider embeddings/vectordb and browser automation last.

