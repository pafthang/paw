# Go Core Stage 2 — Chat

This stage adds the first minimal LLM chat path to the Go core.

## Added

- `internal/llm` package with a small common interface
- Ollama chat client using `POST /api/chat`
- OpenAI-compatible chat client using `POST /v1/chat/completions`
- `paw chat` / `paw ask` CLI command
- `POST /api/v1/chat` API endpoint
- config keys:
  - `model`
  - `ollama_host`
  - `openai_compatible_base_url`
  - `openai_api_key`

## Try Ollama

```bash
go run ./cmd/paw config set agent_backend ollama
go run ./cmd/paw config set ollama_host http://127.0.0.1:11434
go run ./cmd/paw config set model qwen2.5:7b

go run ./cmd/paw chat "Say hello from Paw Go core"
```

## Try OpenAI-compatible

```bash
go run ./cmd/paw config set agent_backend openai_compatible
go run ./cmd/paw config set openai_compatible_base_url https://api.openai.com/v1
go run ./cmd/paw config set openai_api_key "$OPENAI_API_KEY"
go run ./cmd/paw config set model gpt-4o-mini

go run ./cmd/paw chat "Say hello from Paw Go core"
```

## API

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

Or with explicit messages:

```bash
curl -s http://127.0.0.1:8888/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"model":"qwen2.5:7b","messages":[{"role":"user","content":"Say hello"}]}'
```

## Still not included

- streaming responses
- tool calls
- agent loop
- memory/session persistence
- WebSocket chat protocol
- dashboard integration
