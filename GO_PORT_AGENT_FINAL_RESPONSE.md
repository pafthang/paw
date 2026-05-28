# Go Core — Agent Final Response Loop

This stage completes the first practical two-step agent flow:

1. Ask the LLM whether tools are needed.
2. Execute requested tools.
3. Feed tool results back into a second LLM call.
4. Return a final human-readable answer.

## What changed

Previously, `paw agent` could run tool calls, but the model's raw tool-call JSON was the main answer.

Now, if tools are used, Paw sends the tool results back to the LLM and returns `final_response`.

## CLI

```bash
go run ./cmd/paw agent "Read README.md and tell me what this project does."
```

With JSON output:

```bash
go run ./cmd/paw agent --json "Read README.md and tell me what this project does."
```

The JSON response now includes both:

- `model_response`: first LLM response, usually strict JSON tool calls
- `final_response`: second LLM response after tool results are available

## API

```bash
curl -s http://127.0.0.1:8888/api/v1/agent/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "Read README.md and tell me what this project does.",
    "model": "qwen2.5:7b"
  }'
```

## Response shape

```json
{
  "session_id": 1,
  "model_response": {
    "model": "...",
    "content": "{\"tool_calls\":[...]}"
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

## Persistence

For sessions, Paw stores:

- the user prompt
- the final assistant answer when tools were used
- the normal model response when no tools were used

It does not store the raw intermediate tool-call JSON as the assistant's primary session response.

## Next steps

- multiple agent iterations instead of one tool batch
- better JSON recovery for imperfect model output
- allow/deny tool policies
- workspace sandboxing
- provider-native tool calling where supported
