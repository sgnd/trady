# trady

[![CI](https://github.com/sgnd/trady/actions/workflows/ci.yml/badge.svg)](https://github.com/sgnd/trady/actions/workflows/ci.yml)
[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)

**One local command line that turns your long recordings into short captioned clips. For you and your AI agent.**

A long recording holds a few minutes worth sharing. Finding them, cutting
them and adding captions usually means an editor, a web service that wants
your upload, and your own scripts. trady does it on your machine, behind
one command that you and an AI agent (Claude Code, Codex, Cursor, Gemini
CLI, your own bot) can both use:

- It finds the moments from a local transcript and cuts them with ffmpeg.
- Captions are burned in from speech that whisper.cpp turns into text on
  your machine.
- Nothing is uploaded. trady makes no network calls, and its records are
  encrypted (AES-256-GCM) in `~/.trady`.
- Secrets go in a vault. An AI agent sets them from the environment and
  cannot print them.
- Readable text for you, JSON for an agent.
- One file to install: a static Go binary, built like
  [komando](https://github.com/sgnd/komando).

## Install

On macOS or Linux, open a terminal and paste:

```bash
curl --proto '=https' --tlsv1.2 -fsSL https://raw.githubusercontent.com/sgnd/trady/main/install.sh | sh
```

It installs the newest release, and only when the download matches the
release's `checksums.txt`. On Windows, download `trady_windows_amd64.zip`
from [Releases](https://github.com/sgnd/trady/releases) and put
`trady.exe` in a folder on your `PATH`.

Check that it works:

```bash
trady version
trady doctor
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
