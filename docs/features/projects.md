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
- **moment** — scored window (placeholders until local AI scoring)
- **clip** — render result path / skip reason
- **job** — long-running `project.run` tracking

## Honest scope

`project run` selects evenly spaced placeholder moments and cuts with ffmpeg
when installed. Full transcript + LLM scoring is not claimed as production-ready.
