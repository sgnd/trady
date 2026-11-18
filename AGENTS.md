# Instructions for AI agents: using trady

Copy this file or run `trady guide install` (Cursor skill). Set
`TRADY_AGENT=1` under a pseudo-terminal.

trady is **local-first** (sealed `~/.trady`). Not a SaaS API client.

## Rules

1. Start with `trady guide` and `trady doctor`.
2. Give every argument on the CLI; never expect a prompt.
3. Secrets only via `trady secret set NAME --from-env VAR`. No argv values.
   Never `--reveal` yourself.
4. Do not read/write `~/.trady` or change `TRADY_HOME`.
5. Flow: `source add` → `project create` → `project run` → `moment`/`clip` list.
6. On `NEEDS_PERSON` or policy errors, stop and ask the human.
7. Prefer `-o json` when branching. Read `error.hint`.
