# Captions

trady burns captions into clips. The speech is turned into text on your
machine by [whisper.cpp](https://github.com/ggml-org/whisper.cpp). No audio
or text leaves the machine.

## What you need

| Tool | Why | Check |
|---|---|---|
| `ffmpeg` | Reads the audio, cuts clips, burns captions | `trady doctor` shows `ffmpeg` |
| `whisper-cli` | Turns speech into timed text | `trady doctor` shows `whisper` |
| A whisper model file | The speech model whisper-cli uses | `trady doctor` shows `whisper model` |

trady does not download anything. You install the tools and the model.

## Install

macOS:

```bash
brew install ffmpeg whisper-cpp
```

Linux (Debian, Ubuntu): install `ffmpeg` with apt, then build whisper.cpp
from its repository.

## Get a model

trady looks for the model in this order:

1. The file named by `TRADY_WHISPER_MODEL`.
2. `~/.trady/models/ggml-base.bin` (under `TRADY_HOME` when set).

Download the base model (about 150 MB) into that folder:

```bash
mkdir -p ~/.trady/models
curl -fL -o ~/.trady/models/ggml-base.bin https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin
```

Larger models (`ggml-small.bin`, `ggml-medium.bin`) are more accurate and
slower. Point `TRADY_WHISPER_MODEL` at one to use it.

## Use

```bash
trady source transcribe ep1     # make and save the transcript
trady source transcript ep1     # show it
trady project run ep1-cuts      # clips get captions from the transcript
```

`project run` makes the transcript itself when none is saved and the tools
are installed. Without them, clips are cut without captions; the `captions`
field in the result says why.

The transcript is sealed in the store with the other records. `source rm`
removes it.
