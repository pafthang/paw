# Go Core — WebSocket Chat and Agent Events

This stage turns `/ws` from a simple hello/echo compatibility endpoint into a JSON event channel for chat and agent operations.

## Endpoint

```text
/ws?access_token=<token>
```

Token helper:

```bash
TOKEN=$(go run ./cmd/paw auth token)
```

## Hello event

On connect, Paw sends:

```json
{
  "type": "hello",
  "service": "paw",
  "commands": ["chat", "agent.chat", "ping", "echo"],
  "time": "..."
}
```

## Ping

Request:

```json
{
  "id": "1",
  "type": "ping"
}
```

Response:

```json
{
  "id": "1",
  "type": "pong",
  "time": "..."
}
```

## Chat

Request:

```json
{
  "id": "chat-1",
  "type": "chat",
  "prompt": "Say hello",
  "model": "qwen2.5:7b"
}
```

Events:

```json
{"id":"chat-1","type":"chat.started"}
{"id":"chat-1","type":"chat.result","response":{"model":"...","content":"..."}}
```

Current note: websocket `chat` is a lightweight LLM call and does not yet persist sessions. Use `agent.chat` for session-aware agent runs.

## Agent chat

Request:

```json
{
  "id": "agent-1",
  "type": "agent.chat",
  "prompt": "Read README.md and tell me what this project does.",
  "model": "qwen2.5:7b",
  "session_id": 1
}
```

Events:

```json
{"id":"agent-1","type":"agent.started","session_id":1}
{"id":"agent-1","type":"agent.tools","tool_calls":[...],"tool_run_response":{...}}
{"id":"agent-1","type":"agent.result","response":{...}}
```

If no tools are used, Paw skips `agent.tools` and sends only `agent.result`.

## Error events

Errors are returned as:

```json
{
  "id": "agent-1",
  "type": "agent.error",
  "error": "...",
  "time": "..."
}
```

## Next steps

- persist websocket `chat` sessions
- stream partial model output when provider supports streaming
- emit finer-grained agent events: `tool.started`, `tool.result`, `tool.error`
- add session subscriptions
