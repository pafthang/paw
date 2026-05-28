# Go Core Stage 6 — Context Budget

This stage adds a small context-packing layer before LLM calls.

## Added

- `internal/contextpack` package
- default system prompt
- rough context budget by character count
- CLI flags:
  - `--system`
  - `--max-context-chars`
- API fields:
  - `system_prompt`
  - `max_context_chars`
- responses now include context stats:
  - message count
  - rough character count

## CLI

Continue a session with default context packing:

```bash
go run ./cmd/paw chat --session 1 "Continue with the current context."
```

Use a custom system prompt:

```bash
go run ./cmd/paw chat --session 1 \
  --system "You are Paw, a concise Go migration assistant." \
  "What should we do next?"
```

Limit packed context roughly by characters:

```bash
go run ./cmd/paw chat --session 1 \
  --history-limit 50 \
  --max-context-chars 8000 \
  "Summarize the session so far."
```

JSON output includes context stats:

```bash
go run ./cmd/paw chat --json --session 1 "What do you remember?"
```

Example shape:

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

## API

```bash
curl -s http://127.0.0.1:8888/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": 1,
    "history_limit": 50,
    "max_context_chars": 8000,
    "system_prompt": "You are Paw, a concise Go migration assistant.",
    "prompt": "Summarize the session so far."
  }'
```

## Notes

This is a deliberately simple budget. It uses rough character counts, not tokenizer-specific token counts. That is enough to prevent unbounded context growth while keeping the next steps simple.

Future stages can add:

- tokenizer-aware budgeting
- session summaries
- memory extraction
- per-model context defaults
- persisted system prompts per session
