# Local store and secrets

Data directory: `~/.trady` (override with `TRADY_HOME`), mode 0700.

| File | Role |
|------|------|
| `store.db` | Sealed-row store (AES-256-GCM via `sealrow` / `sealdb`; pure Go file backend) |
| `store.key` | Master key when not using keychain or passphrase |
| `exports/` | Default clip export directory under the home |

## Key modes

trady picks the first that applies when it creates the store:

| Mode | When | Notes |
|---|---|---|
| `passphrase` | `TRADY_VAULT_PASS` is set | At least 15 characters; the key comes from scrypt |
| `keychain` | macOS, and `TRADY_KEYCHAIN` is not `off` | The key lives in the Keychain |
| `keyfile` | Otherwise (CI, Linux, Windows) | The key lives in `store.key` |

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
