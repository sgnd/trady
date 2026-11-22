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

`trady doctor` prints the exact next command for each missing tool on your
system.

## Install

macOS:

```bash
brew install ffmpeg whisper-cpp
```

Linux (Debian, Ubuntu): install `ffmpeg` with apt, then build whisper.cpp
and put `whisper-cli` on your `PATH`:

```bash
sudo apt install ffmpeg
git clone https://github.com/ggml-org/whisper.cpp
cd whisper.cpp
cmake -B build
cmake --build build -j --config Release
sudo cp build/bin/whisper-cli /usr/local/bin/
```

Windows: `winget install Gyan.FFmpeg`, then download `whisper-bin-x64.zip` from
the [whisper.cpp releases](https://github.com/ggml-org/whisper.cpp/releases)
and put `whisper-cli.exe` in a folder on your `PATH`. trady is built for
Windows but has not been run there yet.

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

## How the captions look

Captions are word by word. Each line shows up to four words, and every word
fills with colour while it is spoken: white text that turns yellow, with a
black outline, at the bottom of the picture. The size follows the video, so
vertical and horizontal clips both read well. `source transcribe` asks
whisper.cpp for word times and keeps them with the transcript.

A transcript made before v0.1.0 has line times only. Its clips get one
caption line per spoken segment; run `trady source transcribe` again to get
word by word. Words in a recording cannot restyle the captions: style
blocks such as `{\an8}` are dropped.

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
