# Security

trady will hold local project state and secrets (API keys for optional
models, paths under your home). A security bug can expose that data to an
AI agent or another user of the same machine.

## Report a problem

Please don't open a public issue for a security problem. Report it privately:

1. Open the [Security tab](https://github.com/sgnd/trady/security) of this repo.
2. Click **Report a vulnerability**.

Include what you found, how to reproduce it, and what an attacker could do.
You get an answer within a week. Once it is fixed, a new release says what
changed.

## What will count (as the store lands)

- Reading a vault secret without a human approving it (JSON output, errors, logs).
- An AI agent seeing secret values that should stay `[hidden]` / `secret:NAME`.
- Tampering with `~/.trady/store.db` that goes unnoticed.
- Changing `TRADY_HOME` for an agent without a person's yes.

## Right now (Phase 0)

This scaffolding only creates `~/.trady` (mode 0700). There is no sealed
store and no vault yet. Still: do not put secrets in the home folder by
hand; wait for `trady secret`.
