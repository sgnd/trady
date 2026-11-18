# Privacy

trady makes no network calls. Your recordings, transcripts, clips and
secrets stay on your machine.

## What runs

| Program | What it does | Network |
|---|---|---|
| `trady` | Keeps records in the sealed store under `~/.trady` | None: the binary has no network code |
| `ffprobe`, `ffmpeg` | Read and cut local media files | trady gives them local paths only |
| `whisper-cli` | Turns local audio into text with a local model file | None |
| `/usr/bin/security` (macOS) | Keeps the store key in the keychain | None |

`trady doctor` shows each tool it found, the whisper model, and the
`network` line.

## How this stays true

A test (`cmd/trady/privacy_test.go`) fails when any network package gets
into the binary. A future feature that needs the network, such as the
paid cloud sync, must be opt-in and must change that test on purpose.

## What trady does not do

- It does not download tools or models. You install them
  (see [captions.md](captions.md)).
- It does not send usage data.
