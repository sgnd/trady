# Security

trady holds local project state and secrets (sealed store under `~/.trady`,
vault for API keys and similar). A security bug can expose that data to an
AI agent or another user of the same machine.

## Report a problem

Please don't open a public issue for a security problem. Report it privately:

1. Open the [Security tab](https://github.com/sgnd/trady/security) of this repo.
2. Click **Report a vulnerability**.

Include what you found, how to reproduce it, and what an attacker could do.
You get an answer within a week. Once it is fixed, a new release says what
changed.

## What counts

- Reading a vault secret without a human approving it (JSON output, errors, logs).
- An AI agent seeing secret values that should stay hidden / `secret:NAME`.
- Tampering with `~/.trady` that goes unnoticed.
- Changing `TRADY_HOME` for an agent without a person's yes.

## Supported versions

Only the newest release gets fixes. Prefer install from Releases (or
`install.sh`) when available, not an untagged `main` build, for production use.
