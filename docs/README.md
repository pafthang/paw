# Paw Go Core — документация

Эта папка собирает разрозненные заметки по Go-порту Paw в понятную документацию.

## Что здесь есть

| Файл | О чем |
|---|---|
| [01-overview.md](01-overview.md) | Общая картина проекта и поэтапный прогресс |
| [02-cli-config-server.md](02-cli-config-server.md) | CLI, конфиг, сервер, health/status |
| [03-llm-providers.md](03-llm-providers.md) | Ollama, OpenAI-compatible, Anthropic |
| [04-sessions-context-memory.md](04-sessions-context-memory.md) | Сессии, история, replay context, memory roadmap |
| [05-agent-tools-loop.md](05-agent-tools-loop.md) | Agent loop, tool calls, file/shell tools, final response |
| [06-api-websocket.md](06-api-websocket.md) | HTTP API, auth, WebSocket events |
| [07-security-audit-sandbox.md](07-security-audit-sandbox.md) | Token auth, workspace sandbox, shell policy, audit log |
| [08-roadmap.md](08-roadmap.md) | Дорожная карта Stage 5–8 |
| [99-source-map.md](99-source-map.md) | Откуда взяты исходные заметки |

## Быстрый запуск

```bash
go run ./cmd/paw config init
go run ./cmd/paw db init
go run ./cmd/paw doctor
go run ./cmd/paw serve --host 127.0.0.1 --port 8888
```

Проверка API:

```bash
curl -s http://127.0.0.1:8888/api/v1/health
curl -s http://127.0.0.1:8888/api/v1/status
```

## Текущее состояние

Go core уже описан как side-by-side реализация рядом с существующим Python-кодом. Реализованы базовые CLI/API поверхности, LLM chat, провайдеры, сессии, tool execution foundation, agent loop, контекстный replay и context budget. Следующий крупный блок — полноценная память, file store/search, улучшение сессий и дальнейшая доводка agent/runtime слоя.
