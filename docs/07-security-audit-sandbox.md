# Security, sandbox и audit

## Token auth

Protected API endpoints require the existing access token.

Token location:

```text
~/.pocketpaw/access_token
```

Accepted auth forms:

```http
Authorization: Bearer <token>
X-Paw-Access-Token: <token>
```

For WebSocket, token can also be passed as query parameter:

```text
/ws?access_token=<token>
```

## Workspace sandboxing

File tools are restricted to the configured workspace root.

Denied examples when workspace is repo root:

```text
../outside.txt
/etc/passwd
~/.ssh/id_rsa
```

Allowed examples:

```text
README.md
internal/server/server.go
tmp/paw-test.txt
```

## Shell policy

`shell.run` is denied unless shell is explicitly allowed.

CLI flags:

```text
--allow-shell
--allow-shell-dangerous
```

Policy:

- `shell.run` denied by default;
- dangerous shell commands denied unless `allow_shell_dangerous=true`;
- denials are audited as `type=tool.denied`.

## Audit log

GORM model:

```text
AuditEvent
```

Each tool run stores:

- optional session id;
- event type;
- tool name;
- input JSON;
- output JSON;
- error text;
- timestamp.

CLI:

```bash
go run ./cmd/paw audit list
```

API:

```bash
curl -s http://127.0.0.1:8888/api/v1/audit?limit=50
```

## Agent safety defaults

Recommended defaults:

```text
max_iterations = 4
workspace = current directory
allow_shell = false
allow_shell_dangerous = false
```

This gives the agent useful file access inside the project while preventing accidental access outside the workspace and preventing shell execution unless the caller explicitly opts in.
