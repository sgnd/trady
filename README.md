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

Clips need `ffmpeg` (it brings `ffprobe`), and captions need `whisper-cli`
and a model. Without them moments still save, and clips stay `skipped`.
`trady doctor` prints the exact command for each missing tool. trady never
installs anything itself:

| System | ffmpeg | whisper-cli | Model |
|---|---|---|---|
| macOS | `brew install ffmpeg` | `brew install whisper-cpp` | `mkdir -p ~/.trady/models && curl -fL -o ~/.trady/models/ggml-base.bin https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin` |
| Linux | `sudo apt install ffmpeg` | build whisper.cpp from its repository ([steps](docs/features/captions.md)) | the same `curl` as on macOS |
| Windows | `winget install Gyan.FFmpeg` | `whisper-bin-x64.zip` from [whisper.cpp releases](https://github.com/ggml-org/whisper.cpp/releases) | `curl.exe -fL -o %USERPROFILE%\.trady\models\ggml-base.bin` and the same URL |

Untested on Windows so far: trady is built for Windows but has not been run there.

More detail: [Install](docs/features/install.md).

## Quick start

```bash
trady source add ./episode.mp4 --name ep1
trady source transcribe ep1
trady project create ep1 --name ep1-cuts
trady project run ep1-cuts --out ./exports
trady moment list ep1-cuts
trady clip list ep1-cuts
```

Agents: set `TRADY_AGENT=1`, prefer `-o json`, start with `trady guide`.

To choose the moments yourself (or let your agent choose them), read the
transcript with `trady source transcript ep1`, then:

```bash
trady moment set ep1-cuts --start 30 --end 62.5 --title "Why the plan failed"
trady project run ep1-cuts --out ./exports   # cuts exactly your picks
trady moment set ep1-cuts --clear            # back to scored moments
```

## Connect your AI agent

```bash
trady guide install            # Claude Code, Codex, Cursor, Gemini CLI
trady guide install --project  # only for the folder you are in
trady guide install --dry-run  # show where it would write
```

It writes the trady skill to `~/.claude/skills/trady/SKILL.md` (Claude Code)
and `~/.agents/skills/trady/SKILL.md` (Codex, Cursor, Gemini CLI and others),
and nothing else. It makes no network call. Run it again after an update: a
changed file is kept as `.bak`. A person runs it; an agent that tries gets
`NEEDS_PERSON`, except with `--dry-run`. See [AGENTS.md](AGENTS.md) for
agents that read no skills.

## Commands

| Area | Commands |
|------|----------|
| Meta | `version`, `doctor`, `guide [install]`, `schema`, `list` |
| Secrets | `secret set\|list\|show\|rm` |
| Media | `source add\|list\|transcribe\|transcript\|rm` |
| Projects | `project create\|list\|get\|status\|run\|rm` |
| Results | `moment list\|set`, `clip list`, `job list\|get` |

## Scope

With a transcript, `project run` ranks moments with simple local rules
(cue words, questions, numbers, speech pace), not an AI judge. Without one
it uses evenly spaced windows. Moments you set with `moment set` replace
the scored ones. Clips are cut with ffmpeg, and captions are burned in word
by word when whisper.cpp and a model are installed: up to four words a line,
each word fills with colour while it is spoken.

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
