# CLI, config и server

## CLI entrypoint

Основная точка входа:

```bash
go run ./cmd/paw --help
```

Базовые команды:

```bash
go run ./cmd/paw help
go run ./cmd/paw status
go run ./cmd/paw doctor
go run ./cmd/paw health
```

## Config

Go core сохраняет совместимость с существующей конфигурацией PocketPaw:

```text
~/.pocketpaw/config.json
```

Команды:

```bash
go run ./cmd/paw config init
go run ./cmd/paw config show
go run ./cmd/paw config path
go run ./cmd/paw config dir
go run ./cmd/paw config set model qwen2.5:7b
```

Важные config keys:

```text
agent_backend
model
ollama_host
openai_compatible_base_url
openai_api_key
anthropic_api_key
```

## Database

Локальная SQLite база хранится рядом с конфигом:

```text
~/.pocketpaw/paw.db
```

Команды:

```bash
go run ./cmd/paw db init
go run ./cmd/paw db path
```

Основные модели:

- `ChatSession`;
- `ChatMessage`;
- `MemoryItem`;
- `AuditEvent`.

## Server

Запуск API сервера:

```bash
go run ./cmd/paw serve --host 127.0.0.1 --port 8888
```

Публичные endpoints:

```text
GET /
GET /api/v1/health
GET /api/v1/status
```

Protected endpoints требуют access token:

```text
GET /api/v1/settings
POST /api/v1/chat
GET /api/v1/sessions
GET /api/v1/sessions/:id
DELETE /api/v1/sessions/:id
GET /api/v1/tools
POST /api/v1/agent/run
POST /api/v1/agent/chat
GET /api/v1/audit
```

## Health/status check

```bash
curl -s http://127.0.0.1:8888/api/v1/health
curl -s http://127.0.0.1:8888/api/v1/status
```
