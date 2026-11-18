# Local store and secrets

Data directory: `~/.trady` (override with `TRADY_HOME`), mode 0700.

| File | Role |
|------|------|
| `store.db` | Sealed-row store (AES-256-GCM via `sealrow` / `sealdb`; pure Go file backend) |
| `store.key` | Master key when not using keychain or passphrase |
| `exports/` | Default clip export directory under the home |

## Key modes

1. **passphrase** — `TRADY_VAULT_PASS` (≥15 chars) at first create; scrypt.
2. **keychain** — macOS Keychain when available (`TRADY_KEYCHAIN=off` skips).
3. **keyfile** — `store.key` (CI / fallback).

`trady doctor` prints `key_mode`.

## Secrets

```bash
trady secret set openai --from-env OPENAI_API_KEY
trady secret list
trady secret show openai          # existence only
trady secret show openai --reveal # person at TTY only
trady secret rm openai
```

Secret values are sealed again under HKDF label `trady vault entries v1`.
