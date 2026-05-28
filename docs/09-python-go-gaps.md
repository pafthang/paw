# Python → Go gap analysis

Этот документ заменяет корневой `GO_PORT_PYTHON_GAPS.md` и хранит gap analysis внутри `docs/`.

Go core сейчас является focused side-by-side реализацией рядом с большой Python-версией в `src/pocketpaw/`. Цель Go-порта — не слепо переписать весь Python surface, а стабилизировать компактное ядро: CLI, API compatibility, agent/tooling, storage, skills/MCP и channels.

## Что Go уже покрывает

| Область | Статус | Комментарий |
|---|---:|---|
| CLI core | `[done]` | `serve`, `chat`, `agent`, `status`, `doctor`, `config`, `auth`, `db`, `sessions`, `memory`, `file-store`, `search`, `skills`, `mcp`, `channels`, `tools`, `run-tool`, `audit` |
| API compatibility subset | `[done]` | health/status/settings/chat/sessions/tools/agent/audit + memory/files/search + skills + mcp + channels + `/ws` |
| LLM providers | `[done]` | Ollama, OpenAI-compatible, OpenAI, Anthropic |
| Agent loop | `[done]` | multi-iteration loop, tool calls, final response, workspace policy, audit |
| Storage | `[done]` | SQLite `~/.pocketpaw/paw.db`, file store `~/.pocketpaw/files` |
| Sessions | `[done]` | list/show/search/rename/delete, context replay/budget |
| Memory | `[done]` | CRUD/search via CLI and API |
| File store | `[done]` | import/list/show/search/delete via CLI and API |
| Global search | `[partial]` | Basic local search over memory/sessions/messages/files; no embeddings yet |
| Skills | `[done]` | YAML format, load/list/show/validate/reload, install/uninstall, agent injection |
| MCP | `[partial]` | Config/process manager and presets exist; full MCP tool protocol integration can be incremental |
| Channels | `[partial]` | Core manager + Telegram adapter; Discord/Slack deferred |
| Channel audit | `[done]` | Telegram emits receive/send/error audit events |

## Major Python areas not yet ported

### API surface beyond Go compatibility subset

Python `src/pocketpaw/api/v1/*` contains many routers that are not part of the current Go subset.

Status: `[todo]`, but do not blindly port.

Potential areas:

- OAuth2 / scoped auth: `oauth2.py`, `api_keys.py`, `identity.py`, `auth.py`.
- Integrations / connectors / kits: `connectors.py`, `kits.py`, `oauth_integrations.py`.
- Operational/telemetry: `metrics.py`, `analytics.py`, `traces.py`, `events.py`, `alerts.py`.
- Reminders / proactive / plan-mode: `reminders.py`, `plan_mode.py`, related `daemon/*`.
- Remote tunnel / remote control: `remote.py`.
- Soul / deep work / mission control: `soul.py`, `deep_work/*`, `mission_control/*`.

Recommended approach:

1. First define which endpoints are actually required by the Go dashboard/UI/client.
2. Port only the minimum stable contract.
3. Keep auth and policy decisions explicit before adding broad remote/connector surfaces.

### OAuth2 / scoped auth / API keys

Status: `[todo]`

Go currently keeps a simpler access-token model. Python has broader auth concepts that may be needed later for UI clients, cloud/EE, or external integrations.

Recommended next step:

- Write a design doc before implementation.
- Decide whether Go should keep local `access_token` as the primary mode or introduce scoped tokens/API keys.
- Avoid mixing OAuth, connector secrets and local CLI auth in one large PR.

### Integrations / connectors / kits

Status: `[todo]`

Python has a richer integrations surface. Go should only port this after core policy/auth is clear.

Recommended next step:

- Inventory which connectors the UI/client actually calls.
- Add a small `internal/integrations` interface if needed.
- Keep secrets out of audit logs and config dumps.

### Metrics / analytics / traces / events / alerts

Status: `[todo]`

Useful for operational visibility, but not required for local Go core MVP.

Recommended next step:

- Start with local structured events/audit first.
- Add metrics only after there is a stable server runtime story.

### Reminders / proactive / daemon / scheduler

Status: `[todo]`

Python has proactive/daemon concepts. In Go this should be explicit and opt-in.

Recommended design:

- Separate package such as `internal/daemon` or `internal/scheduler`.
- Start via explicit CLI command, not silently from normal chat/agent commands.
- Reuse existing audit and tool policy.
- Persist scheduled jobs in SQLite if they need to survive restarts.

### Vector DB / embeddings

Status: `[todo]`

Python has `vectordb/*` and retrieval-related code. Go currently has basic local search.

Recommended order:

1. Keep SQL/basic search until relevance is a real blocker.
2. Add a compact embedding abstraction.
3. Prefer an embedded/local option first.
4. Treat Chroma or external vector DB as optional adapter, not mandatory dependency.

### Browser automation

Status: `[todo]`

Python `browser/*` implies a larger Playwright/Selenium-like subsystem.

Recommended approach:

- Port only when explicitly needed.
- Browser automation must go through the same sandbox/policy/audit model as shell/file tools.
- Avoid enabling it by default.

### Additional channels

Status: `[deferred]`

Telegram exists. Discord and Slack are not priority unless a real use case appears.

Potential future packages:

```text
internal/channels/discord
internal/channels/slack
```

Recommended approach:

- Add fake-adapter tests first.
- Keep per-channel authorization rules explicit.
- Keep all message send/receive/error events audited.

### Soul protocol / A2A / mission control / deep work

Status: `[deferred]`

These areas can balloon the architecture. They should come after storage, policy, channels, and auth are stable.

Recommended approach:

- Do not port wholesale.
- Extract small primitives only when the Go product direction requires them.
- Keep orchestration/fabric concepts out of core packages until interfaces are clear.

### EE / cloud / fleet / pawprint / fabric

Status: `[deferred]`

EE should remain docs/design-first until core is stable.

Boundary rules:

- Core builds without EE-specific services.
- EE depends on core interfaces, not the reverse.
- No hard-coded vendor endpoints or credentials.
- Build tags, if used, must be documented.

## Recommended porting order from current state

1. `[high]` Add/verify tests for Stage 5–7 implemented surfaces.
2. `[high]` Stabilize auth design: local access token vs scoped API keys/OAuth2.
3. `[medium]` Audit which Python API endpoints are required by the current UI/dashboard/client.
4. `[medium]` Fill only required Go API gaps.
5. `[medium]` Design daemon/scheduler if reminders/proactive flows become important.
6. `[low]` Add embeddings/vector search after basic search is insufficient.
7. `[low]` Add browser automation only behind explicit policy controls.
8. `[low]` Revisit Discord/Slack only after Telegram is proven stable.
9. `[deferred]` Keep Soul/A2A/mission-control/EE/fabric as later architecture work.

## Do not port yet

Avoid porting these until there is a concrete caller/use case:

- Full OAuth/provider matrix.
- Full connector marketplace or kits system.
- Browser automation runtime.
- Mission-control/fabric orchestration.
- EE cloud/fleet runtime.
- External vector DB requirement.

## Practical next checklist

```bash
go test ./...

paw status
paw doctor
paw memory add fact "The project is called Paw."
paw memory search Paw
paw file-store list
paw sessions list
paw search Paw
paw skills list
paw mcp presets
paw channels status
```

Then verify HTTP:

```bash
TOKEN=$(paw auth token)

curl -s http://127.0.0.1:8888/api/v1/status \
  -H "Authorization: Bearer $TOKEN"

curl -s http://127.0.0.1:8888/api/v1/memory \
  -H "Authorization: Bearer $TOKEN"

curl -s "http://127.0.0.1:8888/api/v1/search?q=Paw" \
  -H "Authorization: Bearer $TOKEN"

curl -s http://127.0.0.1:8888/api/v1/channels/status \
  -H "Authorization: Bearer $TOKEN"
```
