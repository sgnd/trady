# Install

## From source

```bash
git clone https://github.com/sgnd/trady.git
cd trady
make build
./bin/trady version
```

## install.sh

When GitHub Releases exist:

```bash
curl --proto '=https' --tlsv1.2 -fsSL https://raw.githubusercontent.com/sgnd/trady/main/install.sh | sh
```

Until then, `install.sh` prints build-from-source instructions for `latest`.

## Dependencies

| Tool | Needed for |
|---|---|
| The trady binary | Everything; it is pure Go (`CGO_ENABLED=0`) and needs nothing else |
| `ffmpeg`, `ffprobe` (optional) | Reading media and cutting clips |
| `whisper-cli` and a model (optional) | Captions and moment scoring; see [captions.md](captions.md) |
