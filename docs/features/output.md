# Output

Every command prints either:

- text in a terminal (tables and lines for people), or
- one JSON envelope when stdout is not a TTY, or with `-o json` / `TRADY_OUTPUT=json`.

```json
{
  "ok": true,
  "command": "source list",
  "data": { },
  "meta": { "contract": 1, "version": "dev", "duration_ms": 12 }
}
```

Agents must set `TRADY_AGENT=1` if stdin looks like a TTY so trady never prompts.
Errors carry `code`, `message`, and `hint`.
