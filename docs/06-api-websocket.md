# API и WebSocket

## HTTP API summary

Public endpoints:

```text
GET /
GET /api/v1/health
GET /api/v1/status
```

Protected endpoints:

```text
GET    /api/v1/settings
POST   /api/v1/chat
GET    /api/v1/sessions
GET    /api/v1/sessions/:id
DELETE /api/v1/sessions/:id
GET    /api/v1/tools
POST   /api/v1/agent/run
POST   /api/v1/agent/chat
GET    /api/v1/audit
```

Planned memory endpoints:

```text
GET    /api/v1/memory
POST   /api/v1/memory
GET    /api/v1/memory/:id
DELETE /api/v1/memory/:id
GET    /api/v1/memory/search?q=...
```

## Access token

Token path:

```text
~/.pocketpaw/access_token
```

CLI:

```bash
go run ./cmd/paw auth token
go run ./cmd/paw auth path
go run ./cmd/paw auth rotate
```

Protected `/api/*` endpoints accept either:

```http
Authorization: Bearer <token>
```

or:

```http
X-Paw-Access-Token: <token>
```

Example:

```bash
TOKEN=$(go run ./cmd/paw auth token)

curl -s http://127.0.0.1:8888/api/v1/settings   -H "Authorization: Bearer $TOKEN"
```

## WebSocket endpoint

Endpoint:

```text
GET /ws
```

Token can be passed via:

```text
Authorization: Bearer <token>
X-Paw-Access-Token: <token>
?access_token=<token>
```

Example URL:

```text
ws://127.0.0.1:8888/ws?access_token=<token>
```

Initial implementation:

- route exists;
- token auth works;
- WebSocket upgrade works;
- clients can send and receive JSON;
- compatibility hello/echo loop exists before full streaming protocol.

## Agent chat events

Top-level events:

```text
agent.started
agent.result
agent.error
```

Iteration-scoped events:

```text
agent.iteration.started
agent.iteration.model_result
agent.iteration.finished
```

Tool progress events include `iteration`:

```text
agent.tool.started
agent.tool.result
agent.tool.error
agent.tool.denied
```

Example:

```json
{"type":"agent.iteration.started","id":"agent-1","iteration":1,"time":"..."}
```

## Next protocol steps

- streaming chat events;
- session-scoped websocket messages;
- agent progress events;
- audit/tool events over WebSocket;
- dashboard integration.
