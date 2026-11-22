---
name: trady
description: Local-first CLI for short captioned clips from long recordings. Use whenever the task involves trady, local media sources, clip projects, moments, or ~/.trady.
---

# trady

trady turns long local recordings into short clip candidates on your machine.
State is sealed under `~/.trady`. It is not a SaaS thin client.

## Start

1. `trady guide`: how to use this version.
2. `trady doctor`: home, store key mode, ffmpeg, whisper-cli.
3. `trady list`: sources and projects.
4. `trady schema COMMAND` when unsure.

Install:
`curl --proto '=https' --tlsv1.2 -fsSL https://raw.githubusercontent.com/sgnd/trady/main/install.sh | sh`
Set `TRADY_AGENT=1` under a pseudo-terminal so trady never prompts.

## Commands

- `trady version`: print the version
- `trady doctor`: check TRADY_HOME, the store, ffmpeg, whisper-cli and the model, and print the next command for each missing tool
- `trady guide`: print the playbook
- `trady guide install`: teach Claude Code, Codex, Cursor and Gemini CLI to use trady (a person runs it; an agent may add --dry-run)
- `trady schema`: print the usage line of every command
- `trady list`: list sources and projects
- `trady secret set`: store a secret in the vault
- `trady secret list`: list secret names
- `trady secret show`: show a secret (a person, with --reveal)
- `trady secret rm`: remove a secret
- `trady source add`: add a recording
- `trady source list`: list recordings
- `trady source transcribe`: make a transcript with word times, locally
- `trady source transcript`: show the transcript
- `trady source rm`: remove a recording
- `trady project create`: make a clip project for a recording
- `trady project list`: list projects
- `trady project get`: show a project
- `trady project status`: show a project's progress
- `trady project run`: pick moments, cut the clips and burn in word-by-word captions
- `trady project rm`: remove a project
- `trady moment list`: list the moments of a project
- `trady moment set`: pick the moments yourself (or clear the picks)
- `trady clip list`: list the clips of a project
- `trady job list`: list jobs
- `trady job get`: show one job

## Rules

1. Give every argument on the command line. Never expect a prompt.
2. Secrets: `trady secret set NAME --from-env VAR`. Never argv values. Never `--reveal` as an agent.
3. Do not read or edit `~/.trady` or change `TRADY_HOME`.
4. Flow: `source add` → `source transcribe` → `project create` → `project run` → `moment list` / `clip list`.
   To choose the moments yourself: read `source transcript`, then `trady moment set PROJECT --start SEC --end SEC --title TEXT` for each pick (5 to 180 seconds), then `project run` again. It cuts exactly your picks. `moment set PROJECT --clear` goes back to scoring.
5. With a transcript, `project run` ranks moments locally and burns word-by-word captions in; without one it uses evenly spaced windows. A missing tool or model is for the human to install: `trady doctor` prints the exact command for each, and you show it to them. Do not run installs yourself.
6. On `NEEDS_PERSON` / policy stops, ask the human. Do not retry with other flags on your own.
7. Prefer `-o json` when branching on results.
8. Run `trady guide install` only when the human asks, and it is theirs to run: as an agent you get `NEEDS_PERSON`. `--dry-run` shows where it would write.
