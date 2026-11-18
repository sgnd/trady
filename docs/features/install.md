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

- **Required:** none beyond the trady binary (pure Go, `CGO_ENABLED=0`).
- **Optional:** `ffmpeg` / `ffprobe` on `PATH` for media probe and clip export.
