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

WebSocket `chat` uses the same session persistence and context packing path as the HTTP chat endpoint.

Create a new session:

```json
{
  "id": "chat-1",
  "type": "chat",
  "prompt": "Say hello",
  "model": "qwen2.5:7b"
}
```

Continue an existing session:

```json
{
  "id": "chat-2",
  "type": "chat",
  "session_id": 1,
  "history_limit": 20,
  "max_context_chars": 8000,
  "prompt": "Continue from the previous answer."
}
```

Events:

```json
{"id":"chat-1","type":"chat.started","session_id":0}
{"id":"chat-1","type":"chat.result","session_id":1,"history_messages":0,"context":{"messages":2,"chars":1234},"response":{"model":"...","content":"..."}}
```

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

- stream partial model output when provider supports streaming
- emit finer-grained agent events: `tool.started`, `tool.result`, `tool.error`
- add session subscriptions
