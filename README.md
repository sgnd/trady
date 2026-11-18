# trady

[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/github/go-mod/go-version/sgnd/trady)](go.mod)
[![Release](https://img.shields.io/github/v/release/sgnd/trady?include_prereleases&sort=semver)](https://github.com/sgnd/trady/releases)

**Local-first CLI for short, captioned clips from long recordings. For you and your AI agent.**

Long files sit on disk. Editors and scripts each invent their own folders,
metadata and secrets. trady keeps sources, projects, moments and clips
behind one command that you and an AI agent (Claude Code, Codex, Cursor,
Gemini CLI, your own bot) can both use.

- Everything stays on your machine. Records are encrypted (AES-256-GCM) in
  `~/.trady`, and trady makes no network calls.
- Secrets go in a vault. Agents set them from the environment and cannot
  print them.
- You get readable text; an agent gets JSON.
- It is built like [komando](https://github.com/sgnd/komando): one static
  Go binary (`CGO_ENABLED=0`), no prompts for agents, and `guide` and
  `schema` commands for them.

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
clip export work, and `whisper-cli` with a model for captions. Without them, moments still save; clips stay `skipped`.

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
| Meta | `version`, `doctor`, `guide [install]`, `schema`, `list` |
| Secrets | `secret set\|list\|show\|rm` |
| Media | `source add\|list\|transcribe\|transcript\|rm` |
| Projects | `project create\|list\|get\|status\|run\|rm` |
| Results | `moment list`, `clip list`, `job list\|get` |

## Scope

With a transcript, `project run` ranks moments with simple local rules
(cue words, questions, numbers, speech pace), not an AI judge. Without one
it uses evenly spaced windows. Clips are cut with ffmpeg, and captions are
burned in when whisper.cpp and a model are installed. Captions are one line
per spoken segment, not word by word.

## Docs

- [Features](docs/features/README.md)
- [Install](docs/features/install.md)
- [Store](docs/features/store.md)
- [Projects](docs/features/projects.md)
- [Captions](docs/features/captions.md)
- [Privacy](docs/features/privacy.md)
- [Output](docs/features/output.md)
- [Changelog](CHANGELOG.md)

## License

[AGPL-3.0](LICENSE). Commercial dual-license: [COMMERCIAL-LICENSE.md](COMMERCIAL-LICENSE.md).
Contributors: [CLA.md](CLA.md).

© Sugeng Agung Suganda
