# Go Core Stage 2 — Auth and WebSocket Compatibility

This stage fills the missing compatibility pieces from the original Stage 2 plan:

- access token path: `~/.pocketpaw/access_token`
- API auth middleware
- `/ws` websocket endpoint
- CLI helpers for token management

## Access token

Generate or print the local token:

```bash
go run ./cmd/paw auth token
```

Print token path:

```bash
go run ./cmd/paw auth path
```

Rotate token:

```bash
go run ./cmd/paw auth rotate
```

The token is stored at:

```text
~/.pocketpaw/access_token
```

## HTTP auth

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

curl -s http://127.0.0.1:8888/api/v1/settings \
  -H "Authorization: Bearer $TOKEN"
```

Public endpoints:

```text
GET /
GET /api/v1/health
GET /api/v1/status
```

## WebSocket

The websocket endpoint is available at:

```text
/ws
```

It requires the same access token, either as a query parameter:

```text
ws://127.0.0.1:8888/ws?access_token=<token>
```

or via header where the client supports custom websocket headers.

The initial implementation is a compatibility endpoint with a hello message and echo loop. It establishes the `/ws` surface for dashboard/API parity before adding streaming chat messages.

## Notes

This is not yet the full Python websocket protocol. It is the Go compatibility surface:

- route exists
- token auth works
- websocket upgrade works
- clients can send and receive JSON

Next steps:

- streaming chat events
- session-scoped websocket messages
- agent progress events
- audit/tool events over websocket
