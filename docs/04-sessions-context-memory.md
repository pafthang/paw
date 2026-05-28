# Sessions, context replay и memory roadmap

## Session persistence

Stage 4 добавляет durable chat history через SQLite/GORM.

CLI:

```bash
go run ./cmd/paw chat "Say hello and remember this session"
go run ./cmd/paw chat --session 1 "Continue that answer"
go run ./cmd/paw sessions list
go run ./cmd/paw sessions show 1
go run ./cmd/paw sessions delete 1
```

API:

```bash
curl -s http://127.0.0.1:8888/api/v1/chat   -H 'Content-Type: application/json'   -d '{"prompt":"Say hello from Paw Go core"}'
```

Append to session:

```bash
curl -s http://127.0.0.1:8888/api/v1/chat   -H 'Content-Type: application/json'   -d '{"session_id":1,"prompt":"Continue"}'
```

List/show/delete:

```bash
curl -s http://127.0.0.1:8888/api/v1/sessions?limit=20
curl -s http://127.0.0.1:8888/api/v1/sessions/1
curl -X DELETE -s http://127.0.0.1:8888/api/v1/sessions/1
```

## Context replay

Stage 5 делает сохраненные сессии полезными как контекст.

CLI:

```bash
go run ./cmd/paw chat "My project is called Paw and we are porting it to Go."
go run ./cmd/paw chat --session 1 "What is my project called?"
go run ./cmd/paw chat --session 1 --history-limit 6 "Summarize the session so far."
go run ./cmd/paw chat --session 1 --history-limit 0 "Start a fresh thought in this session."
```

API:

```bash
curl -s http://127.0.0.1:8888/api/v1/chat   -H 'Content-Type: application/json'   -d '{"session_id":1,"history_limit":6,"prompt":"Summarize the session so far."}'
```

Response включает:

```text
session_id
history_messages
response
```

## Context budget

Stage 6 добавляет context packing перед LLM call.

CLI:

```bash
go run ./cmd/paw chat --session 1 "Continue with the current context."

go run ./cmd/paw chat --session 1   --system "You are Paw, a concise Go migration assistant."   "What should we do next?"

go run ./cmd/paw chat --session 1   --history-limit 50   --max-context-chars 8000   "Summarize the session so far."
```

JSON response содержит context stats:

```json
{
  "session_id": 1,
  "history_messages": 8,
  "context": {
    "messages": 10,
    "chars": 4200
  },
  "response": {
    "model": "...",
    "content": "..."
  }
}
```

API:

```bash
curl -s http://127.0.0.1:8888/api/v1/chat   -H 'Content-Type: application/json'   -d '{
    "session_id": 1,
    "history_limit": 50,
    "max_context_chars": 8000,
    "system_prompt": "You are Paw, a concise Go migration assistant.",
    "prompt": "Summarize the session so far."
  }'
```

## Memory roadmap

Следующий логичный блок — полноценная память и file store layer.

Планируемые CLI команды:

```bash
paw memory list
paw memory show <id>
paw memory add <type> <content>
paw memory delete <id>
paw memory search <query>
```

Планируемые API routes:

```text
GET    /api/v1/memory
POST   /api/v1/memory
GET    /api/v1/memory/:id
DELETE /api/v1/memory/:id
GET    /api/v1/memory/search?q=...
```

Ожидаемая JSON shape:

```json
{
  "type": "fact",
  "content": "The project is called Paw.",
  "metadata": {
    "source": "manual"
  }
}
```
