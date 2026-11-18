# Plan

The one list of planned work for trady. Deferred items stay here, marked
"deferred" with the reason and what they wait on.

Order: the CLI first, until it is solid. Then the paid plan
(`efolusi/trady-apps`, `efolusi/trady-cloud`), built the way komando's apps
and cloud are built.

## CLI

| # | Milestone | Status |
|---|---|---|
| 1 | End-to-end CLI test; name conflicts (`CONFLICT`, exit 8); empty lists are `[]`; `project rm` removes its moments and clips; weak vault key gives `KEY` (exit 6) | done |
| 2 | Unit tests for `inventory`, `out`, `media` (moment placement) and `store` (key modes) | done |
| 3 | Argument checks: reject extra arguments and unknown flags in every command | open |
| 4 | Real clip export with ffmpeg: cut, captions burned in, test with a generated sample video in CI | open |
| 5 | Local moment scoring (replace the evenly spaced placeholders) | open |
| 6 | `secret` and vault: tests for set, list, show, rm; agent can never print a value | open |
| 7 | Releases: goreleaser, checksums, `install.sh` against real Releases | open; tag and release need the owner's yes |
| 8 | Docs: features pages match every command; `trady schema` matches the code | open |

## Paid plan

| # | Milestone | Status |
|---|---|---|
| 1 | `efolusi/trady-cloud`: API, account, sync, billing, website, following komando-cloud | deferred: waits on CLI milestones 1 to 8 |
| 2 | `efolusi/trady-apps`: Mac, iPhone, iPad app, following komando-apps | deferred: waits on CLI milestones 1 to 8 |
