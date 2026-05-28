# Overview: Go-порт Paw

## Цель

Go core переносит стабильные runtime-поверхности Paw/PocketPaw в Go, не удаляя существующие Python-директории `src/` и `ee/`. Подход — постепенный side-by-side порт: сначала CLI, конфиг и API, затем LLM, сессии, tools, agent loop, память и WebSocket/dashboard совместимость.

## Реализованные этапы

### Stage 1 — базовый Go core

Реализовано:

- `cmd/paw` как Go entrypoint;
- `paw serve` API-only сервер;
- `paw status`;
- `paw doctor` / `paw health`;
- `paw config show/init/path/dir/set`;
- совместимость с `~/.pocketpaw/config.json`;
- совместимость с `~/.pocketpaw/access_token`;
- базовые JSON API routes:
  - `GET /`;
  - `GET /api/v1/health`;
  - `GET /api/v1/status`;
  - `GET /api/v1/settings`.

### Stage 2 — chat и auth/WebSocket совместимость

Добавлено:

- минимальный LLM chat path;
- Ollama provider;
- OpenAI-compatible provider;
- `paw chat` / `paw ask`;
- `POST /api/v1/chat`;
- access token middleware;
- `/ws` WebSocket endpoint;
- CLI helpers для token management.

### Stage 3 — application stack и LLM providers

Принят стек:

- Cobra для CLI;
- Echo для HTTP/WebSocket API;
- GORM + SQLite для локальной персистентности.

LLM providers:

- `ollama`;
- `openai_compatible`, aliases: `openai-compatible`, `openai`;
- `anthropic`, alias: `claude`.

### Stage 4 — sessions, tools, agent loop

Добавлено:

- durable chat history через SQLite;
- `paw sessions list/show/delete`;
- file tools: `file.read`, `file.write`;
- shell tool: `shell.run`;
- audit log;
- LLM-driven tool calls;
- multi-step agent loop;
- workspace sandboxing;
- tool allow/deny policy;
- final response loop после выполнения tools.

### Stage 5 — session context replay

Сохраненные сессии начали использоваться как контекст, а не только как история:

- `paw chat --session` подгружает последние сообщения;
- `--history-limit` ограничивает replay;
- API принимает `session_id` и `history_limit`;
- response показывает `history_messages`.

### Stage 6 — context budget

Добавлен простой context-packing слой:

- `internal/contextpack`;
- default system prompt;
- rough char-based context budget;
- CLI флаги `--system`, `--max-context-chars`;
- API поля `system_prompt`, `max_context_chars`;
- context stats в ответе.

## Что еще не закрыто полностью

- tokenizer-aware budgeting;
- session summaries;
- memory extraction;
- file store/search layer;
- полноценная memory API/CLI;
- dashboard integration;
- provider-native tool calling;
- более надежный JSON recovery для imperfect model output.
