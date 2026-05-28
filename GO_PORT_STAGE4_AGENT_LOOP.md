# Go Core Stage 4 — Agent Loop Foundation

This stage fills in the missing Stage 4 foundation that should sit after chat sessions:

- tool calls
- file read/write tools
- shell tool
- audit log
- minimal agent runner

The existing session/chat/context work remains in place, but this PR names the actual agent-loop foundation explicitly.

## Added packages

```text
internal/agent
internal/tools
```

## Tools

Available tools:

```text
file.read   Read a UTF-8 text file from disk.
file.write  Write UTF-8 text content to disk.
shell.run   Run a shell command when explicitly allowed.
```

`shell.run` requires `allow=true` in its input JSON.

## Audit log

A new SQLite/GORM model is added:

```text
AuditEvent
```

Each tool run writes an audit event with:

- optional session id
- event type
- tool name
- input JSON
- output JSON
- error text
- timestamp

## CLI

List tools:

```bash
go run ./cmd/paw tools
```

Read a file:

```bash
go run ./cmd/paw run-tool file.read \
  --input '{"path":"README.md"}'
```

Write a file:

```bash
go run ./cmd/paw run-tool file.write \
  --input '{"path":"tmp/paw-test.txt","content":"hello"}'
```

Run shell command:

```bash
go run ./cmd/paw run-tool shell.run \
  --input '{"command":"go test ./...","allow":true,"timeout_seconds":60}'
```

List audit events:

```bash
go run ./cmd/paw audit list
```

## API

List tools:

```bash
curl -s http://127.0.0.1:8888/api/v1/tools
```

Run tools:

```bash
curl -s http://127.0.0.1:8888/api/v1/agent/run \
  -H 'Content-Type: application/json' \
  -d '{
    "tool_calls": [
      {"name":"file.read","input":{"path":"README.md"}}
    ]
  }'
```

Run shell:

```bash
curl -s http://127.0.0.1:8888/api/v1/agent/run \
  -H 'Content-Type: application/json' \
  -d '{
    "tool_calls": [
      {"name":"shell.run","input":{"command":"go test ./...","allow":true,"timeout_seconds":60}}
    ]
  }'
```

List audit events:

```bash
curl -s http://127.0.0.1:8888/api/v1/audit?limit=50
```

## Still not included

This is not yet autonomous LLM-driven function calling. It is the safe execution foundation.

Next steps:

- LLM tool-call parsing
- agent planning loop
- tool allow/deny policy
- workspace sandboxing
- richer audit UI/API
