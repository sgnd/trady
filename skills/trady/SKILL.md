---
name: trady
description: Local-first CLI for short captioned clips from long recordings. Use whenever the task involves trady, local media sources, clip projects, moments, or ~/.trady.
---

# trady

trady turns long local recordings into short clip candidates on your machine.
State is sealed under `~/.trady`. It is not a SaaS thin client.

## Start

1. `trady guide` — playbook for this version.
2. `trady doctor` — home, store key mode, ffmpeg, whisper-cli.
3. `trady list` — sources and projects.
4. `trady schema COMMAND` when unsure.

Install: build from source (`make build`) until Releases exist.
Set `TRADY_AGENT=1` under a pseudo-terminal so trady never prompts.

## Rules

1. Give every argument on the command line. Never expect a prompt.
2. Secrets: `trady secret set NAME --from-env VAR`. Never argv values. Never `--reveal` as an agent.
3. Do not read or edit `~/.trady` or change `TRADY_HOME`.
4. Flow: `source add` → (`source transcribe`) → `project create` → `project run` → `moment list` / `clip list`.
5. With a transcript, `project run` ranks moments locally and burns captions in; without one it uses evenly spaced windows. A missing tool or model is for the human to install.
6. On `NEEDS_PERSON` / policy stops, ask the human — do not retry with dangerous flags on your own.
7. Prefer `-o json` when branching on results.
