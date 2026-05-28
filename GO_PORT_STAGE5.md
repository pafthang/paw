# Go Core Stage 5 — Session Context Replay

This stage makes saved chat sessions useful as context, not just as history.

## Added

- `paw chat --session <id>` now loads recent saved session messages before calling the LLM.
- `paw chat --history-limit <n>` controls how many previous messages are replayed.
- `POST /api/v1/chat` with `session_id` now loads recent saved session messages before calling the LLM.
- `POST /api/v1/chat` accepts `history_limit`.
- Responses include `history_messages` so callers can see how much context was replayed.
- `internal/db.ListRecentChatMessages` returns the newest N messages in chronological order.

## CLI

Create a new saved session:

```bash
go run ./cmd/paw chat "My project is called Paw and we are porting it to Go."
```

Continue with context:

```bash
go run ./cmd/paw chat --session 1 "What is my project called?"
```

Limit replayed history:

```bash
go run ./cmd/paw chat --session 1 --history-limit 6 "Summarize the session so far."
```

Disable replayed history while still appending to the session:

```bash
go run ./cmd/paw chat --session 1 --history-limit 0 "Start a fresh thought in this session."
```

## API

Start the server:

```bash
go run ./cmd/paw serve --host 127.0.0.1 --port 8888
```

Create a new session:

```bash
curl -s http://127.0.0.1:8888/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt":"My project is called Paw and we are porting it to Go."}'
```

Continue with context:

```bash
curl -s http://127.0.0.1:8888/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"session_id":1,"prompt":"What is my project called?"}'
```

Limit replayed history:

```bash
curl -s http://127.0.0.1:8888/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"session_id":1,"history_limit":6,"prompt":"Summarize the session so far."}'
```

## Notes

This is still a simple replay strategy. The next steps can add summarization, token budgeting, memory extraction, and configurable system prompts.
