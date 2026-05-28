# Paw Go Core — документация

Эта папка собирает разрозненные заметки по Go-порту Paw в понятную документацию. Корневые roadmap/gap файлы перенесены сюда и структурированы.

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
| [08-roadmap.md](08-roadmap.md) | Сверенный roadmap/status Stage 5–8: что реализовано, что partial/todo/deferred |
| [09-python-go-gaps.md](09-python-go-gaps.md) | Python → Go gap analysis: что еще не портировано и в каком порядке браться |

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

Go core — side-by-side реализация рядом с существующим Python-кодом. Уже реализованы CLI/API поверхности, LLM providers, sessions, memory, file-store, search, skills, MCP, agent/tool loop, audit, WebSocket и Telegram channel baseline.

Главные следующие задачи: тесты для уже реализованных Stage 5–7 поверхностей, auth/OAuth/API-key design, минимальный порт нужных Python API endpoints для UI/client, затем daemon/vector/browser только при реальной необходимости.
