# Instructions for AI agents: using trady

Copy this file, or have the human run `trady guide install`: it adds the
trady skill for Claude Code, Codex, Cursor and Gemini CLI. Set
`TRADY_AGENT=1` under a pseudo-terminal.

trady is **local-first** (sealed `~/.trady`). Not a SaaS API client.

## Rules

1. Start with `trady guide` and `trady doctor`.
2. Give every argument on the CLI; never expect a prompt.
3. Secrets only via `trady secret set NAME --from-env VAR`. No argv values.
   Never `--reveal` yourself.
4. Do not read/write `~/.trady` or change `TRADY_HOME`.
5. Flow: `source add` → `source transcribe` → `project create` → `project run` → `moment`/`clip` list.
   To pick the moments yourself: `moment set PROJECT --start SEC --end SEC
   --title TEXT`, then `project run`; `moment set PROJECT --clear` undoes it.
6. On `NEEDS_PERSON` or policy errors, stop and ask the human.
7. Prefer `-o json` when branching. Read `error.hint`.
8. A missing ffmpeg, whisper-cli or model is the human's to install: show them
   the command `trady doctor` prints. The home-wide `trady guide install` is theirs to run (`--project` and `--dry-run` are yours).

## Commands

- `trady version`: print the version
- `trady doctor`: check TRADY_HOME, the store, ffmpeg, whisper-cli and the model, and print the next command for each missing tool
- `trady guide`: print the playbook
- `trady guide install`: teach Claude Code, Codex, Cursor and Gemini CLI to use trady (a person runs it; an agent may add --project or --dry-run)
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

Run `trady schema COMMAND` for the flags of each.
