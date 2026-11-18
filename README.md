# trady

[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)

**Local-first CLI for short, captioned clips from long recordings.**  
For you and your AI agent.

Trady mengelola **proyek media lokal**: sumber file di mesin Anda, state
tersimpan sealed di `~/.trady` (e2ee), secret di vault. Runtime utama adalah
CLI ini — **bukan** thin client ke trady.efolusi.com.

Arsitektur saudara [komando](https://github.com/sgnd/komando): binary
statis (`CGO_ENABLED=0`), store lokal sealed (e2ee, pure Go), agent-safe,
text + JSON.

## Quick start

```bash
make build
./bin/trady doctor
./bin/trady source add ./episode.mp4 --name ep1
./bin/trady project create ep1 --name ep1-cuts
./bin/trady project run ep1-cuts --out ./exports
./bin/trady moment list ep1-cuts
./bin/trady clip list ep1-cuts
```

Opsional: pasang `ffmpeg`/`ffprobe` agar probe durasi dan export clip jalan.
Tanpa ffmpeg, momen tetap tersimpan; clip berstatus `skipped`.

## Commands

| Area | Commands |
|------|----------|
| Meta | `version`, `doctor`, `guide`, `schema`, `list` |
| Secrets | `secret set\|list\|show\|rm` |
| Media | `source add\|list\|rm` |
| Projects | `project create\|list\|get\|status\|run\|rm` |
| Results | `moment list`, `clip list\|export`, `job list\|get` |

Agent: `TRADY_AGENT=1`, `-o json`. Playbook: `trady guide`.

## Scope (honest)

`project run` memakai **placeholder moments** (jendela merata) lalu memotong
dengan ffmpeg bila ada. Scoring AI lokal / caption word-level belum
production-ready — struktur store dan perintah sudah siap dilanjutkan.

## Install

Dari source: `make build`.  
`install.sh` mengarah ke Releases `sgnd/trady` (placeholder sampai rilis pertama).

## License

[AGPL-3.0](LICENSE). Commercial dual-license: [COMMERCIAL-LICENSE.md](COMMERCIAL-LICENSE.md).
Contributors: [CLA.md](CLA.md).

© Sugeng Agung Suganda
