# trady guide

trady is a local-first CLI: long recordings → short captioned-clip workflow on your machine.
State is sealed under `~/.trady` (override `TRADY_HOME`). It is not a thin client to trady.efolusi.com.

## For AI agents

1. Run `trady doctor`, then `trady list`.
2. Never prompt: pass every flag. Set `TRADY_AGENT=1` under a pseudo-terminal.
3. Prefer `-o json` when branching on results.
4. Secrets: `trady secret set NAME --from-env VAR`. Never put values on the argv. `secret show --reveal` is for people only.
5. Do not read or rewrite `~/.trady` by hand.

## Typical flow

```bash
trady source add ./episode.mp4 --name ep1
trady project create ep1 --name ep1-cuts
trady project run ep1-cuts --out ./exports
trady moment list ep1-cuts
trady clip list ep1-cuts
```

`project run` currently builds evenly spaced placeholder moments and cuts them with ffmpeg when installed. Local AI scoring is not production-ready yet.

## Safety

- Agent never sees vault values (`[hidden]`).
- Destructive or reveal actions that need a person stop with `NEEDS_PERSON`.
