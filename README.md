# trady

[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/github/go-mod/go-version/sgnd/trady)](go.mod)
[![Release](https://img.shields.io/github/v/release/sgnd/trady?include_prereleases&sort=semver)](https://github.com/sgnd/trady/releases)

**Local-first CLI for short, captioned clips from long recordings. For you and your AI agent.**

Long files sit on disk. Editors and scripts each invent their own folders,
metadata and secrets. trady keeps sources, projects, moments and clips
behind one command that you and an AI agent (Claude Code, Codex, Cursor,
Gemini CLI, your own bot) can both use:

- **Sources and projects stay on your machine.** State lives sealed in
  `~/.trady` (e2ee). This is the runtime, not a thin client to a cloud API.
- **Secrets stay in a vault.** Agents set them from the environment; they
  do not print them.
- **Readable text for you, JSON for an agent.**
- **Same shape as [komando](https://github.com/sgnd/komando):** static Go
  binary (`CGO_ENABLED=0`), agent-safe CLI, guide and schema for agents.

## Install

From source (until the first GitHub Release):

```bash
git clone https://github.com/sgnd/trady.git
cd trady
make build
./bin/trady version
./bin/trady doctor
```

When Releases exist:

```bash
curl --proto '=https' --tlsv1.2 -fsSL https://raw.githubusercontent.com/sgnd/trady/main/install.sh | sh
```

Optional: put `ffmpeg` / `ffprobe` on your `PATH` so duration probe and
clip export work. Without them, moments still save; clips stay `skipped`.

More detail: [Install](docs/features/install.md).

## Quick start

```bash
./bin/trady source add ./episode.mp4 --name ep1
./bin/trady project create ep1 --name ep1-cuts
./bin/trady project run ep1-cuts --out ./exports
./bin/trady moment list ep1-cuts
./bin/trady clip list ep1-cuts
```

Agents: set `TRADY_AGENT=1`, prefer `-o json`, start with `trady guide`.

## Commands

| Area | Commands |
|------|----------|
| Meta | `version`, `doctor`, `guide`, `schema`, `list` |
| Secrets | `secret set\|list\|show\|rm` |
| Media | `source add\|list\|rm` |
| Projects | `project create\|list\|get\|status\|run\|rm` |
| Results | `moment list`, `clip list\|export`, `job list\|get` |

## Scope (honest)

`project run` uses **placeholder moments** (even windows) then cuts with
ffmpeg when available. Local AI scoring and word-level captions are not
production-ready yet; the store and commands are ready to extend.

## Docs

- [Features](docs/features/README.md)
- [Store](docs/features/store.md)
- [Projects](docs/features/projects.md)
- [Output](docs/features/output.md)

## License

[AGPL-3.0](LICENSE). Commercial dual-license: [COMMERCIAL-LICENSE.md](COMMERCIAL-LICENSE.md).
Contributors: [CLA.md](CLA.md).

© Sugeng Agung Suganda
