# LLM providers

Go core использует общий `internal/llm` interface и несколько provider implementations.

## Ollama

Backend:

```text
ollama
```

Используется endpoint:

```text
POST /api/chat
```

Настройка:

```bash
go run ./cmd/paw config set agent_backend ollama
go run ./cmd/paw config set ollama_host http://127.0.0.1:11434
go run ./cmd/paw config set model qwen2.5:7b
```

Проверка:

```bash
go run ./cmd/paw chat "Say hello from Paw Go core"
```

## OpenAI-compatible

Backend:

```text
openai_compatible
```

Aliases:

```text
openai-compatible
openai
```

Используется endpoint:

```text
POST /v1/chat/completions
```

Настройка:

```bash
go run ./cmd/paw config set agent_backend openai_compatible
go run ./cmd/paw config set openai_compatible_base_url https://api.openai.com/v1
go run ./cmd/paw config set openai_api_key "$OPENAI_API_KEY"
go run ./cmd/paw config set model gpt-4o-mini
```

Проверка:

```bash
go run ./cmd/paw chat "Say hello from Paw Go core"
```

## Anthropic

Backend:

```text
anthropic
```

Alias:

```text
claude
```

Используется Messages API:

```text
POST /v1/messages
```

Headers:

```text
x-api-key
anthropic-version
```

Настройка:

```bash
go run ./cmd/paw config set agent_backend anthropic
go run ./cmd/paw config set anthropic_api_key "$ANTHROPIC_API_KEY"
go run ./cmd/paw config set model claude-3-5-haiku-latest
```

Особенность: internal `system` messages конвертируются в top-level Anthropic `system` text, а остальные сообщения отправляются как `user` / `assistant`.

## Chat API

```bash
curl -s http://127.0.0.1:8888/api/v1/chat   -H 'Content-Type: application/json'   -d '{"prompt":"Say hello from Paw Go core"}'
```

С explicit messages:

```bash
curl -s http://127.0.0.1:8888/api/v1/chat   -H 'Content-Type: application/json'   -d '{"model":"qwen2.5:7b","messages":[{"role":"user","content":"Say hello"}]}'
```
