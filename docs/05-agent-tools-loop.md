# Agent, tools и loop

## Tool execution foundation

Stage 4 добавляет основу для безопасного выполнения tools.

Пакеты:

```text
internal/agent
internal/tools
```

Доступные tools:

```text
file.read   Read a UTF-8 text file from disk.
file.write  Write UTF-8 text content to disk.
shell.run   Run a shell command when explicitly allowed.
```

`shell.run` требует явного разрешения.

## CLI tools

List tools:

```bash
go run ./cmd/paw tools
```

Read file:

```bash
go run ./cmd/paw run-tool file.read   --input '{"path":"README.md"}'
```

Write file:

```bash
go run ./cmd/paw run-tool file.write   --input '{"path":"tmp/paw-test.txt","content":"hello"}'
```

Run shell:

```bash
go run ./cmd/paw run-tool shell.run   --input '{"command":"go test ./...","allow":true,"timeout_seconds":60}'
```

## API tools

List tools:

```bash
curl -s http://127.0.0.1:8888/api/v1/tools
```

Run file read:

```bash
curl -s http://127.0.0.1:8888/api/v1/agent/run   -H 'Content-Type: application/json'   -d '{
    "tool_calls": [
      {"name":"file.read","input":{"path":"README.md"}}
    ]
  }'
```

Run shell:

```bash
curl -s http://127.0.0.1:8888/api/v1/agent/run   -H 'Content-Type: application/json'   -d '{
    "tool_calls": [
      {"name":"shell.run","input":{"command":"go test ./...","allow":true,"timeout_seconds":60}}
    ]
  }'
```

## LLM-driven tool calls

Agent prompt просит модель вернуть strict JSON, если нужны tools:

```json
{
  "tool_calls": [
    {
      "name": "file.read",
      "input": {
        "path": "README.md"
      }
    }
  ]
}
```

Если tools не нужны, модель отвечает обычным текстом.

CLI:

```bash
go run ./cmd/paw agent "Read README.md and tell me what this project does."
go run ./cmd/paw agent --json "Read README.md and tell me what this project does."
go run ./cmd/paw agent --session 1 "Now inspect go.mod too."
go run ./cmd/paw agent --model qwen2.5:7b "Read README.md."
```

API:

```bash
curl -s http://127.0.0.1:8888/api/v1/agent/chat   -H 'Content-Type: application/json'   -d '{
    "prompt": "Read README.md and tell me what this project does.",
    "model": "qwen2.5:7b"
  }'
```

With session:

```bash
curl -s http://127.0.0.1:8888/api/v1/agent/chat   -H 'Content-Type: application/json'   -d '{
    "session_id": 1,
    "prompt": "Now inspect go.mod too."
  }'
```

## Final response loop

Практический двухшаговый flow:

1. спросить LLM, нужны ли tools;
2. выполнить requested tools;
3. передать tool results во второй LLM call;
4. вернуть финальный human-readable answer.

Раньше `paw agent` мог вернуть raw tool-call JSON как основной ответ. Теперь, если tools используются, Paw отправляет результаты обратно в LLM и возвращает `final_response`.

JSON response включает:

```json
{
  "session_id": 1,
  "model_response": {
    "model": "...",
    "content": "{"tool_calls":[...]}"
  },
  "final_response": {
    "model": "...",
    "content": "This project is..."
  },
  "tool_calls": [
    {
      "name": "file.read",
      "input": {
        "path": "README.md"
      }
    }
  ],
  "tool_run_response": {
    "results": [
      {
        "tool_name": "file.read",
        "result": {
          "content": "..."
        }
      }
    ]
  },
  "used_tools": true
}
```

## Multi-step agent loop

Stage 4 считается complete, когда agent loop поддерживает:

- multi-step tool/LLM iterations, default `max_iterations=4`;
- allow/deny policy;
- workspace sandboxing;
- audit events для allowed и denied tool calls.

CLI flags:

```text
--max-iterations N
--workspace /path/to/project
--allow-shell
--allow-shell-dangerous
```

HTTP request:

```json
{
  "prompt": "Read README.md and summarize.",
  "max_iterations": 4,
  "workspace": "/path/to/project",
  "allow_shell": false
}
```
