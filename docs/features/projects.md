# Media projects

## Flow

```bash
trady source add ./long.mp4 --name ep1
trady project create ep1 --name ep1-cuts
trady project run ep1-cuts --out ./exports
trady moment list ep1-cuts
trady clip list ep1-cuts
trady job list
```

## Records (sealed)

- **source** — local media path + ffprobe fields when available
- **project** — links to a source; status lifecycle
- **moment** — scored window of the source
- **clip** — render result path / skip reason
- **transcript** — timed text of a source, made locally (see [captions.md](captions.md))
- **job** — long-running `project.run` tracking

## Honest scope

With a transcript, `project run` scores moments locally: it builds 15 to 60
second windows on sentence edges and ranks them by cue words, questions,
numbers and how lively the speech is. It needs no model and is a simple
ranking, not an AI judge. Without a transcript it falls back to evenly
spaced placeholders. The `selection` field in the result says which.

Each run replaces the moments and clips of the last run. Clips are cut with
ffmpeg when installed. With whisper.cpp and a model, captions are burned in. Full transcript + LLM scoring is not claimed as production-ready.
