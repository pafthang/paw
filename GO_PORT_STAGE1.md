# Go Core Stage 1

This branch adds the first Go implementation layer for PocketPaw/Paw without removing or replacing the existing Python code in `src/` and `ee/`.

## Scope

Implemented:

- `cmd/paw` Go entrypoint
- `paw serve` API-only server
- `paw status`
- `paw doctor` / `paw health`
- `paw config show/init/path/dir/set`
- compatibility with the existing `~/.pocketpaw/config.json` location
- compatibility with the existing `~/.pocketpaw/access_token` path convention
- basic JSON API routes:
  - `GET /`
  - `GET /api/v1/health`
  - `GET /api/v1/status`
  - `GET /api/v1/settings`

Not implemented yet:

- LLM chat loop
- WebSocket chat protocol
- dashboard asset serving
- memory/session APIs
- skills/MCP
- channel adapters
- `ee` enterprise features

## Try it

```bash
go run ./cmd/paw help
go run ./cmd/paw config init
go run ./cmd/paw doctor
go run ./cmd/paw serve --host 127.0.0.1 --port 8888
```

Then open:

```text
http://127.0.0.1:8888/api/v1/status
http://127.0.0.1:8888/api/v1/health
```

## Design goal

The Go port starts as a side-by-side core implementation. The current Python package remains intact while the Go core gradually gains parity with the stable runtime surfaces first: config, health, server lifecycle, API routes, agent loop, tools, memory, skills, MCP, channels, and finally `ee`.
