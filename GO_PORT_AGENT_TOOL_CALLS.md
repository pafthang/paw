# Go Core — LLM-driven Tool Calls

This stage connects the existing agent tool execution foundation to the LLM path.

## Added

- `internal/agent/chat.go`
- strict JSON tool-call protocol prompt
- tool-call parser
- LLM-driven agent chat runner
- CLI command:
  - `paw agent <prompt>`
- API route:
  - `POST /api/v1/agent/chat`

## Tool-call protocol

When the LLM needs tools, it should return strict JSON only:

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

If no tools are needed, it can answer normally in plain text.

## CLI examples

Ask the model to inspect a file:

```bash
go run ./cmd/paw agent "Read README.md and tell me what this project does."
```

Force JSON output from Paw:

```bash
go run ./cmd/paw agent --json "Read README.md and tell me what this project does."
```

Continue an existing session:

```bash
go run ./cmd/paw agent --session 1 "Now inspect go.mod too."
```

Use a specific model:

```bash
go run ./cmd/paw agent --model qwen2.5:7b "Read README.md."
```

## API example

```bash
curl -s http://127.0.0.1:8888/api/v1/agent/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "Read README.md and tell me what this project does.",
    "model": "qwen2.5:7b"
  }'
```

With a session:

```bash
curl -s http://127.0.0.1:8888/api/v1/agent/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id": 1,
    "prompt": "Now inspect go.mod too."
  }'
```

## Response shape

```json
{
  "session_id": 1,
  "model_response": {
    "model": "...",
    "content": "..."
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

## Notes

This is the first simple planning loop. It currently performs one LLM turn and one batch of tool calls.

Next steps:

- feed tool results back into a second LLM turn
- add allow/deny policies
- workspace sandboxing
- structured OpenAI/Ollama tool calling where providers support it
- better parser recovery for non-strict model output
