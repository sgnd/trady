# Changelog

All notable changes are listed here, newest first. Versions follow
[Semantic Versioning](https://semver.org), and the format follows
[Keep a Changelog](https://keepachangelog.com).

Below 1.0.0, a patch release (`0.0.x`) has only fixes, tests and docs, and
a minor release (`0.x.0`) adds a feature or changes existing use. Many
changes go into one release; not every merge gets a tag.

## [Unreleased]

The first release will be 0.0.1.

### Added

- Sealed local store under `~/.trady` (`TRADY_HOME`), with three key modes:
  passphrase (`TRADY_VAULT_PASS`), keychain, or key file.
- Vault for secrets: `trady secret set NAME --from-env VAR`, plus `list`,
  `show` and `rm`. A value on the command line is refused, and an AI agent
  cannot reveal a value.
- Media flow: `trady source add|list|rm`, `trady project create|list|get|status|run|rm`,
  `trady moment list`, `trady clip list`, `trady job list|get`, `trady list`.
- `trady source transcribe` and `trady source transcript`: local
  transcription with whisper.cpp. `trady project run` burns captions into
  each clip when a transcript exists or can be made.
- `trady project run` picks moments from the transcript: 15 to 60 second
  windows on sentence edges, ranked locally by cue words, questions,
  numbers and speech pace. Without a transcript it uses evenly spaced
  placeholders; the `selection` field says which.
- trady makes no network calls. `trady doctor` says so in a new `network`
  field, and a test keeps network code out of the binary.
- `trady doctor`, `trady guide`, `trady schema`, and `-o json` output with a
  stable envelope for AI agents.

### Fixed

- A source or project name that is already taken fails with `CONFLICT`
  (exit 8) instead of making a second record.
- Empty lists are `[]` in JSON, not `null`.
- `trady project rm` also removes the project's moments and clips.
- Running `trady project run` again replaces the last run's moments and
  clips instead of adding more.
- A vault passphrase shorter than 15 characters fails with `KEY` (exit 6).
- `trady secret list` with no secrets gives `[]`, not `null`.
- `trady secret set` hides a value typed at a terminal.
- Secret names are 1 to 64 letters, digits, `_`, `-` or `.`; others fail
  with `USAGE`.
- Every command rejects unknown flags, flags without a value, and missing
  or extra arguments (`USAGE`, exit 2).
