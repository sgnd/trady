# Plan

The one list of planned work for trady. Deferred items stay here, marked
"deferred" with the reason and what they wait on.

Order: the CLI first, until it is solid. Then the paid plan
(`efolusi/trady-apps`, `efolusi/trady-cloud`), built the way komando's apps
and cloud are built.

## Goal

A premium killer feature from the internet that runs locally, for people
and AI agents: turn a long recording into short, captioned, ready-to-post
clips, the way paid web clip tools do, without uploading anything.

| Principle | What it means for trady | Where in this plan |
|---|---|---|
| Open source, for transparency | The CLI is AGPL-3.0 and does the whole job alone | CLI 1 to 13 |
| Self-hosted, for ownership | Files, models and the store stay on the user's machine; the cloud can be self-hosted | CLI 10, Paid 1 |
| Private, for protection | No network call unless the user turns one on; models run locally | CLI 10, 11 |
| Secure, for trust | Sealed store, vault for secrets, audits, signed releases | CLI 6, 7, 12 |
| Cloud, for convenience | Optional end-to-end encrypted sync between the CLI and the apps | Paid 1, 2 |
| Commercial, for sustainability | Paid apps and cloud; commercial licence beside AGPL | Paid 1, 2, 3 |

## CLI

| # | Milestone | Status |
|---|---|---|
| 1 | End-to-end CLI test; name conflicts (`CONFLICT`, exit 8); empty lists are `[]`; `project rm` removes its moments and clips; weak vault key gives `KEY` (exit 6) | done |
| 2 | Unit tests for `inventory`, `out`, `media` (moment placement) and `store` (key modes) | done |
| 3 | Argument checks: reject extra arguments and unknown flags in every command | done |
| 4 | Real clip export with ffmpeg: cut, captions burned in, test with a generated sample video in CI | done |
| 5 | Local moment scoring (replace the evenly spaced placeholders) | open |
| 6 | `secret` and vault: tests for set, list, show, rm; agent can never print a value | open |
| 7 | Releases: goreleaser, checksums, `install.sh` against real Releases; first release is 0.0.1, CHANGELOG.md lists what it holds | open; every tag needs the owner's yes |
| 8 | Docs: features pages match every command; `trady schema` matches the code | open |
| 9 | Local transcription (whisper.cpp or similar) for captions and for moment scoring | done for captions; moment scoring is milestone 5 |
| 9a | Run whisper.cpp in CI with the tiny model, so transcription is tested end to end | open; CI would download about 75 MB per run, so it needs a cache |
| 10 | Privacy guarantee: no network call by default; `trady doctor` shows which local models and tools are used | open |
| 11 | Vertical reframe (9:16) with local face or speaker tracking | open |
| 12 | Security audit with the `security-audit` skill; fix what it finds | open |
| 13 | Agent integration: `trady guide install`, an MCP server, and the skill, tested with a real agent | open |

## Paid plan

| # | Milestone | Status |
|---|---|---|
| 1 | `efolusi/trady-cloud`: API, account, sync, billing, website, following komando-cloud | deferred: waits on CLI milestones 1 to 8 |
| 2 | `efolusi/trady-apps`: Mac, iPhone, iPad app, following komando-apps | deferred: waits on CLI milestones 1 to 8 |
| 3 | Pricing, licence checks and checkout for the apps and cloud | deferred: waits on Paid 1 and 2 |
| 4 | Cloud sync is end-to-end encrypted (the cloud never sees media or transcripts) and the cloud can be self-hosted | deferred: part of Paid 1 |
