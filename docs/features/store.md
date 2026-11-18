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
trady secret show openai --reveal # asks the person in a window first
trady secret rm openai
```

Secret values are sealed again under HKDF label `trady vault entries v1`.

`--reveal` asks the person at this computer in a window of their own (macOS,
or Linux with `zenity` and a screen), because a terminal alone cannot tell a
person from an AI agent. Cancel gives `DECLINED`; with no window to show
(SSH, no screen, Windows, `TRADY_NO_DIALOG=1`) it stops with `NEEDS_PERSON`
and the value stays hidden.

On macOS the store key goes to the keychain through `security -i`, so it is
never on a command line.
