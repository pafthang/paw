# Go Core Stage 3 — Application Stack

This stage adopts the Go application stack for the rest of the port:

- `cobra` for CLI commands
- `echo` for HTTP API routing and middleware
- `gorm` for local persistence
- `sqlite` as the default local database

The existing Python implementation in `src/` and `ee/` remains untouched.

## Dependencies

```go
require (
    github.com/labstack/echo/v4 v4.13.4
    github.com/spf13/cobra v1.9.1
    gorm.io/driver/sqlite v1.6.0
    gorm.io/gorm v1.30.0
)
```

## Database

The local SQLite database is stored beside the existing PocketPaw config:

```text
~/.pocketpaw/paw.db
```

Initial GORM models:

- `ChatSession`
- `ChatMessage`
- `MemoryItem`

Initialize/migrate the DB:

```bash
go run ./cmd/paw db init
```

Print DB path:

```bash
go run ./cmd/paw db path
```

## CLI

The CLI is now Cobra-based:

```bash
go run ./cmd/paw --help
go run ./cmd/paw serve --host 127.0.0.1 --port 8888
go run ./cmd/paw status
go run ./cmd/paw doctor
go run ./cmd/paw config show
go run ./cmd/paw chat "hello"
```

## API

The HTTP server is now Echo-based and keeps the existing stage 2 routes:

- `GET /`
- `GET /api/v1/health`
- `GET /api/v1/status`
- `GET /api/v1/settings`
- `POST /api/v1/chat`

Start server:

```bash
go run ./cmd/paw serve --host 127.0.0.1 --port 8888
```

Call chat:

```bash
curl -s http://127.0.0.1:8888/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt":"Say hello from Paw Go core"}'
```

## Why this stack

`cobra` gives a clean path for the existing command surface: `serve`, `status`, `doctor`, `config`, `memory`, `sessions`, `skills`, and future `mcp` commands.

`echo` gives simpler route groups, middleware, WebSocket-friendly integration, and a better fit for gradually recreating the FastAPI dashboard/API surface.

`gorm + sqlite` gives a durable local store for sessions, messages, memory, audit events, skills metadata, and later workspace/project metadata without requiring a server-side database.

## Still not included

- session persistence in chat path
- memory API
- streaming responses
- tool calls
- WebSocket chat protocol
- dashboard integration
