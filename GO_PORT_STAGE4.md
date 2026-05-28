# Go Core Stage 4 — Session Persistence

This stage wires the first durable chat history path into the Go core.

## Added

- Chat exchanges are saved into SQLite via GORM.
- `paw chat` creates a new `ChatSession` by default.
- `paw chat --session <id>` appends to an existing session.
- `paw sessions list`
- `paw sessions show <id>`
- `paw sessions delete <id>`
- `GET /api/v1/sessions`
- `GET /api/v1/sessions/:id`
- `DELETE /api/v1/sessions/:id`
- `POST /api/v1/chat` now returns a `session_id` and persists user/assistant messages.

## Try it

Initialize the DB:

```bash
go run ./cmd/paw db init
```

Run a chat:

```bash
go run ./cmd/paw chat "Say hello and remember this session"
```

Append to an existing session:

```bash
go run ./cmd/paw chat --session 1 "Continue that answer"
```

List sessions:

```bash
go run ./cmd/paw sessions list
```

Show one session:

```bash
go run ./cmd/paw sessions show 1
```

Delete one session:

```bash
go run ./cmd/paw sessions delete 1
```

## API

Start the server:

```bash
go run ./cmd/paw serve --host 127.0.0.1 --port 8888
```

Create a session through chat:

```bash
curl -s http://127.0.0.1:8888/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt":"Say hello from Paw Go core"}'
```

Append to a session:

```bash
curl -s http://127.0.0.1:8888/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"session_id":1,"prompt":"Continue"}'
```

List sessions:

```bash
curl -s http://127.0.0.1:8888/api/v1/sessions?limit=20
```

Show a session:

```bash
curl -s http://127.0.0.1:8888/api/v1/sessions/1
```

Delete a session:

```bash
curl -X DELETE -s http://127.0.0.1:8888/api/v1/sessions/1
```

## Notes

This stage does not yet replay session history back into the LLM request. It persists the exchanges so the next stage can implement contextual session continuation, memory extraction, and dashboard history views.
