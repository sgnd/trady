# Install

## install.sh (macOS, Linux)

```bash
curl --proto '=https' --tlsv1.2 -fsSL https://raw.githubusercontent.com/sgnd/trady/main/install.sh | sh
```

| Setting | What it does |
|---|---|
| `TRADY_VERSION=v0.0.1` | Installs that release instead of the newest |
| `TRADY_INSTALL_DIR=DIR` | Installs into DIR instead of `/usr/local/bin` or `~/.local/bin` |

The script downloads the archive and the release's `checksums.txt`, and
installs nothing when the sha256 does not match. With no release yet, it
says so and shows how to build from source.

Each archive also carries build provenance. Check a download with:

```bash
gh attestation verify trady_linux_amd64.tar.gz --repo sgnd/trady
```

## Windows

Download `trady_windows_amd64.zip` (or `arm64`) from
[Releases](https://github.com/sgnd/trady/releases), check it against
`checksums.txt`, and put `trady.exe` in a folder on your `PATH`.

## From source

```bash
git clone https://github.com/sgnd/trady.git
cd trady
make build
./bin/trady version
```

## Make a release (maintainers)

1. Move the `[Unreleased]` notes in CHANGELOG.md under `## [X.Y.Z] - YYYY-MM-DD`.
2. Push the tag `vX.Y.Z` (it needs the owner's yes).
3. `.github/workflows/release.yml` runs every check, builds `make dist`,
   records build provenance and publishes the release with those notes.

## Dependencies

| Tool | Needed for |
|---|---|
| The trady binary | Everything; it is pure Go (`CGO_ENABLED=0`) and needs nothing else |
| `ffmpeg`, `ffprobe` (optional) | Reading media and cutting clips |
| `whisper-cli` and a model (optional) | Captions and moment scoring; see [captions.md](captions.md) |
