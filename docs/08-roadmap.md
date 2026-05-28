# Roadmap: Stage 5–8

Этот roadmap предназначен для продолжения Go-порта Paw после первых core stages.

## Global rules

- Не трогать существующие Python `src/` и `ee/`, кроме использования как reference.
- Сохранять Go stack:
  - Cobra для CLI;
  - Echo для HTTP/WebSocket API;
  - GORM + SQLite для local persistence.
- Не ломать существующие CLI commands и API routes.
- Делать small focused PRs.
- Добавлять docs и tests где практично.
- Не требовать реальные API keys в тестах.
- Перед сдачей запускать:

```bash
go test ./...
```

---

## Stage 5 — Memory / Sessions

### Goal

Превратить existing session storage в usable memory and file store layer.

Stage 5 должен дать:

```text
paw memory
paw sessions
file_store
search
delete
```

### Baseline

Уже есть:

- `ChatSession`;
- `ChatMessage`;
- `MemoryItem`;
- `paw sessions list/show/delete`;
- SQLite database at `~/.pocketpaw/paw.db`;
- session-aware `paw chat`;
- session-aware `paw agent`;
- session-aware HTTP and WebSocket paths.

### Suggested packages

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

### 5.1 Memory CLI

Required commands:

```bash
paw memory list
paw memory show <id>
paw memory add <type> <content>
paw memory delete <id>
paw memory search <query>
```

Optional aliases:

```bash
paw memory ls
paw memory rm <id>
```

Expected behavior:

- `paw memory add` stores into `MemoryItem`;
- `paw memory list` returns recent memory items;
- `paw memory search` searches content and metadata;
- `paw memory delete` removes a memory item.

JSON output examples:

```bash
paw memory list --json
paw memory search "project name" --json
```

Acceptance criteria:

- CLI commands compile and work;
- memory entries persist in SQLite;
- memory search returns relevant entries;
- deleted memory item disappears from list/search.

### 5.2 Memory API

Required routes:

```text
GET    /api/v1/memory
POST   /api/v1/memory
GET    /api/v1/memory/:id
DELETE /api/v1/memory/:id
GET    /api/v1/memory/search?q=...
```

Protected by existing access token middleware.

Expected JSON shape:

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

- routes require auth token;
- CRUD works;
- search works;
- errors are JSON and actionable.

### 5.3 Session improvements

Extend `paw sessions` with:

```bash
paw sessions search <query>
paw sessions rename <id> <title>
```

Suggested behavior:

- search by title and message content;
- rename updates session title/metadata;
- JSON output should be supported where useful.

---

## Stage 6 — Context and summarization improvements

Continue from rough character budget toward smarter context handling.

Useful next tasks:

- tokenizer-aware budgeting;
- session summaries;
- automatic compact summaries for old context;
- persisted system prompt per session;
- per-model context defaults;
- better response context stats.

---

## Stage 7 — File store and search

Goal: make project/user files first-class context inputs.

Possible features:

- local file metadata table;
- file indexing command;
- text extraction for markdown/text/code files;
- search across indexed files;
- API and CLI for file search;
- optional workspace-scoped indexing.

Suggested commands:

```bash
paw files index <path>
paw files search <query>
paw files list
paw files delete <id>
```

---

## Stage 8 — Agent/runtime polish

Likely work:

- provider-native tool calling where supported;
- stronger JSON recovery for imperfect model output;
- richer tool policies;
- streaming agent progress over WebSocket;
- dashboard views for sessions, tools, memory, audit;
- integration tests for CLI/API flows;
- safer shell/tool UX.
