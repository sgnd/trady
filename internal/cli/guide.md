# trady guide

trady cuts long recordings into short captioned clips, on your machine.
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
trady source transcribe ep1        # optional: needs ffmpeg, whisper-cli and a model
trady project run ep1-cuts --out ./exports
trady moment list ep1-cuts
trady clip list ep1-cuts
```

With a transcript, `project run` ranks moments with local rules and burns captions in. Without one it uses evenly spaced windows and no captions; the `selection` and `captions` fields say which. `trady doctor` shows which tools are installed.

trady makes no network calls and downloads nothing. If a tool or model is missing, tell the human; do not install it yourself.

## Safety

- Agent never sees vault values (`[hidden]`).
- Destructive or reveal actions that need a person stop with `NEEDS_PERSON`.
- `trady schema COMMAND` gives the exact usage lines.
