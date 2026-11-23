# Contributing

Thank you for helping. Please write issues, pull requests and docs in English.

trady is open source under the [AGPL-3.0](LICENSE), and is also sold under a
[commercial license](COMMERCIAL-LICENSE.md). Before your first pull request
is merged, please sign the [Contributor License Agreement](CLA.md): type its
sentence in the pull request description (every commit author signs), and a
check looks for it. You
keep the copyright of what you wrote.

trady is a local-first command line for turning long recordings into short
captioned clips. It follows the same architecture ideas as
[komando](https://github.com/sgnd/komando): one static binary, a sealed
local store, secrets out of sight, safe by default, for people and AI agents.

## Report a bug or suggest a feature

Open an issue on [github.com/sgnd/trady](https://github.com/sgnd/trady).
For a security problem, see [SECURITY.md](SECURITY.md) instead.

## Build and test

You need [Go](https://go.dev/dl/) (the version in [`go.mod`](go.mod) or newer),
`make` and `git`.

```bash
make build    # compile ./bin/trady
make test     # unit tests
```

`make help` lists everything.

## Before you add something

- It fits local media / clip work, or the shared foundation (store, secrets, output).
- It works as a command for a person and for an AI agent: no prompt an agent can meet.
- Prefer no new dependency; if needed, keep it small and permissively licensed (MIT/BSD/Apache). Pure Go (`CGO_ENABLED=0`).

## Changelog

A change people notice gets a line under `[Unreleased]` in
[CHANGELOG.md](CHANGELOG.md), in the same pull request.

## Merging outside pull requests

For maintainers. When a co-author signs by a comment, re-run the CLA check
on the pull request: it runs on pull request events, not on comments.
GitHub maps a commit to a login by its email, which it does not verify; the
person who opened the pull request must still sign, so nothing unsigned is
merged.

Squash-merge a pull request from outside, and add one trailer per signer to
the merge commit, so git keeps the CLA record:

```
CLA-Signed-By: <login> (#<number>, <YYYY-MM-DD>)
```

Never delete the `cla-signatures` branch: it holds the signatures given to
the old bot before 2026-11-23.
